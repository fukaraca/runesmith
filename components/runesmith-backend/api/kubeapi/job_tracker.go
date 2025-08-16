package kubeapi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"
	"github.com/fukaraca/runesmith/shared"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	infv1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/tools/cache"
)

type JobTracker struct {
	node     *config.Meta
	client   *Client
	informer cache.SharedIndexInformer
	stopCh   chan struct{}
	logger   *slog.Logger
	depot    *artifactory.Artifactory

	jobSelector labels.Selector
	jobInf      infv1.JobInformer
}

func NewJobTracker(client *Client, meta *config.Meta, logger *slog.Logger, art *artifactory.Artifactory) (*JobTracker, error) {
	sel, err := labels.Parse("workload-type=enchantment")
	if err != nil {
		return nil, err
	}
	return &JobTracker{
		node:        meta,
		client:      client,
		logger:      logger,
		stopCh:      make(chan struct{}),
		jobSelector: sel,
		depot:       art,
	}, nil
}

func (jw *JobTracker) Start(ctx context.Context) error {
	tweak := func(opts *metav1.ListOptions) {
		opts.LabelSelector = jw.jobSelector.String()
	}

	fac := informers.NewSharedInformerFactoryWithOptions(jw.client.set, 0,
		informers.WithNamespace(jw.client.Namespace),
		informers.WithTweakListOptions(tweak),
	)

	jw.informer = fac.Batch().V1().Jobs().Informer()

	_, err := jw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    jw.onJobAdd,
		UpdateFunc: jw.onJobUpdate,
		DeleteFunc: jw.onJobDelete,
	})
	if err != nil {
		return fmt.Errorf("event handlers could not be add: %w", err)
	}

	go fac.Start(jw.stopCh)

	if ok := cache.WaitForCacheSync(jw.stopCh, jw.informer.HasSynced); !ok {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	jw.logger.Info("job watcher started", slog.String("selector", jw.jobSelector.String()))

	go func() {
		<-ctx.Done()
		jw.Stop()
	}()

	<-jw.stopCh
	return nil
}

func (jw *JobTracker) Stop() {
	select {
	case <-jw.stopCh:
		// already closed
	default:
		close(jw.stopCh)
		jw.logger.Info("job watcher stopped")
	}
}

func (jw *JobTracker) onJobAdd(obj any) {
	j, ok := obj.(*batchv1.Job)
	if !ok {
		return
	}
	// just for action
	jw.logger.Info("job add: job detected", slog.String("name", j.Name))
}

func (jw *JobTracker) onJobUpdate(oldObj, newObj any) {
	j, ok := newObj.(*batchv1.Job)
	if !ok {
		return
	}
	if isComplete(j) {
		jw.depot.MarkArtifactCompleted(string(j.UID), shared.CompletedAS)
	} else if isFailed(j) {
		jw.depot.MarkArtifactCompleted(string(j.UID), shared.FailedAS)
	}
	jw.logger.Info("job update: job detected", slog.String("name", j.Name))
}

func (jw *JobTracker) onJobDelete(obj any) {
	var j *batchv1.Job
	switch o := obj.(type) {
	case *batchv1.Job:
		j = o
	case cache.DeletedFinalStateUnknown:
		if cast, ok := o.Obj.(*batchv1.Job); ok {
			j = cast
		}
	}
	if j == nil {
		return
	}

	if isComplete(j) { // TODO job delete affect on condition needs to be checked
		jw.depot.MarkArtifactCompleted(string(j.UID), shared.CompletedAS) // idempotent already
	} else if isFailed(j) {
		jw.depot.MarkArtifactCompleted(string(j.UID), shared.FailedAS)
	}

	jw.logger.Info("job delete: job detected", slog.String("name", j.Name))
}

func isComplete(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return j.Status.Succeeded > 0
}

func isFailed(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == v1.ConditionTrue {
			return true
		}
	}
	return j.Status.Failed > 0
}
