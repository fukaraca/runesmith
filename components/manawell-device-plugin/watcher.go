package main

import (
	"context"
	"fmt"
	"log/slog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	v1Informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher struct {
	watcher     WatcherConfig
	node        NodeConfig
	manager     *ManaGer
	kubeClients kubernetes.Interface
	informer    cache.SharedIndexInformer
	stopCh      chan struct{}
	logger      *slog.Logger

	podSelector labels.Selector
	resourceKey string

	podInf v1Informer.PodInformer
}

func NewPodWatcher(cfg *Config, manager *ManaGer, logger *slog.Logger) (*PodWatcher, error) {
	clients, err := getKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	sel, err := labels.Parse(fmt.Sprintf("workload-type=enchantment,energy=%s", cfg.Mana.EnergyType))
	if err != nil {
		return nil, fmt.Errorf("parse label selector failed: %w", err)
	}

	return &PodWatcher{
		watcher:     cfg.Watcher,
		node:        cfg.Node,
		manager:     manager,
		kubeClients: clients,
		logger:      logger,
		podSelector: sel,
		resourceKey: fmt.Sprintf("manawell.io/%s", manager.energyType),
		stopCh:      make(chan struct{}),
	}, nil
}

func (pw *PodWatcher) Start(ctx context.Context) error {
	options := []informers.SharedInformerOption{
		informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.FieldSelector = "spec.nodeName=" + pw.node.Name
			lo.LabelSelector = pw.podSelector.String()
		}),
	}
	options = append(options, informers.WithNamespace(pw.node.Namespace))
	fac := informers.NewSharedInformerFactoryWithOptions(
		pw.kubeClients, pw.watcher.ResyncInterval, options...,
	)
	pw.podInf = fac.Core().V1().Pods()
	/* // TODO maybe we reconcile the mana by allocated pod ids
	err := pw.podInf.Informer().AddIndexers(cache.Indexers{
		"byUID": func(obj interface{}) ([]string, error) {
			if pod, ok := obj.(*v1.Pod); ok {
				return []string{string(pod.UID)}, nil
			}
			return []string{}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("add indexer: %w", err)
	}
	*/

	_, err := pw.podInf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pw.onPodAdd,
		UpdateFunc: pw.onPodUpdate,
		DeleteFunc: pw.onPodDelete,
	})
	if err != nil {
		return fmt.Errorf("event handlers could not be add: %w", err)
	}

	go fac.Start(pw.stopCh)

	if ok := cache.WaitForCacheSync(pw.stopCh, pw.podInf.Informer().HasSynced); !ok {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	pw.logger.Info("pod watcher started",
		slog.String("selector", pw.podSelector.String()),
		slog.String("resource", pw.resourceKey),
	)

	go func() {
		<-ctx.Done()
		pw.Stop()
	}()

	<-pw.stopCh
	return nil
}

func (pw *PodWatcher) Stop() {
	select {
	case <-pw.stopCh:
		// already closed
	default:
		close(pw.stopCh)
		pw.logger.Info("pod watcher stopped")
	}
}

func (pw *PodWatcher) onPodAdd(obj any) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	// just for action
	pw.logger.Info("pod add: pod detected", slog.String("name", pod.Name))
}

func (pw *PodWatcher) onPodUpdate(oldObj, newObj any) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}

	if isTerminal(newPod) || newPod.DeletionTimestamp != nil {
		pw.releasePodResources(newPod)
	}
	pw.logger.Info("pod update: pod detected", slog.String("name", newPod.Name))
}

func (pw *PodWatcher) onPodDelete(obj any) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		ghost, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			pw.logger.Warn("pod delete: unexpected type")
			return
		}
		pod, ok = ghost.Obj.(*v1.Pod)
		if !ok {
			pw.logger.Warn("pod delete: obj is not a pod")
			return
		}
	}

	pw.releasePodResources(pod) // it won't affect if it was already released onPodUpdate()
	pw.logger.Info("pod delete: pod detected", slog.String("name", pod.Name))
}

func (pw *PodWatcher) shouldProcessPod(pod *v1.Pod) bool {
	return pod.Spec.NodeName == pw.node.Name && pw.podSelector.Matches(labels.Set(pod.Labels))
}

func isTerminal(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodSucceeded, v1.PodFailed:
		return true
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated != nil {
			return true
		}
	}
	return false
}

func (pw *PodWatcher) releasePodResources(pod *v1.Pod) {
	podID := string(pod.UID)
	if err := pw.manager.ReleaseDevices(podID); err != nil {
		pw.logger.Debug("release failed", slog.String("name", pod.Name), slog.String("uid", podID), slog.Any("err", err))
		return
	}
	pw.logger.Info("released mana", slog.String("name", pod.Name), slog.String("uid", podID))
}

func getKubernetesClient() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}
