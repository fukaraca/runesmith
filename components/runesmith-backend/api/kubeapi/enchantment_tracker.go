package kubeapi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"
	enchantv1 "github.com/fukaraca/runesmith/components/runesmith-operator/api/v1"
	"github.com/fukaraca/runesmith/shared"
	"k8s.io/apimachinery/pkg/api/equality"
	cache2 "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

type EnchantmentTracker struct {
	node     *config.Meta
	client   *Client
	cache    cache.Cache
	informer cache.Informer
	stopCh   chan struct{}
	logger   *slog.Logger
	depot    *artifactory.Artifactory
	ns       string
}

func NewEnchantmentTracker(c *Client, meta *config.Meta, logger *slog.Logger, art *artifactory.Artifactory) (*EnchantmentTracker, error) {
	cc, err := cache.New(c.restConfig, cache.Options{
		Scheme: c.scheme,
		DefaultNamespaces: map[string]cache.Config{
			c.Namespace: {},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create cache: %w", err)
	}

	inf, err := cc.GetInformer(context.Background(), &enchantv1.Enchantment{})
	if err != nil {
		return nil, fmt.Errorf("get enchantment informer: %w", err)
	}

	t := &EnchantmentTracker{
		node:     meta,
		client:   c,
		cache:    cc,
		informer: inf,
		stopCh:   make(chan struct{}),
		logger:   logger,
		depot:    art,
		ns:       c.Namespace,
	}

	_, err = inf.AddEventHandler(cache2.ResourceEventHandlerFuncs{
		AddFunc:    t.onEnchantAdd,
		UpdateFunc: t.onEnchantUpdate,
		DeleteFunc: t.onEnchantDelete,
	})
	if err != nil {
		return nil, fmt.Errorf("event handlers couldn't be added %w", err)
	}

	return t, nil
}

func (t *EnchantmentTracker) Start(ctx context.Context) error {
	go t.cache.Start(ctx) // respects ctx.Done()

	if ok := t.cache.WaitForCacheSync(ctx); !ok {
		return fmt.Errorf("timed out waiting for enchantment cache to sync")
	}
	t.logger.Info("enchantment watcher started", slog.String("namespace", t.ns))

	<-ctx.Done()
	t.logger.Info("enchantment watcher stopping")
	return nil
}

func (t *EnchantmentTracker) Stop() {
	select {
	case <-t.stopCh:
		// already closed
	default:
		close(t.stopCh)
		t.logger.Info("job watcher stopped")
	}
}

func (t *EnchantmentTracker) onEnchantAdd(obj any) {
	e, ok := obj.(*enchantv1.Enchantment)
	if !ok {
		return
	}
	// just action
	t.logger.Info("enchantment add", slog.String("name", e.Name), slog.String("state", e.Status.Phase.String()))
}

func (t *EnchantmentTracker) onEnchantUpdate(oldObj, newObj any) {
	oldE, ok1 := oldObj.(*enchantv1.Enchantment)
	newE, ok2 := newObj.(*enchantv1.Enchantment)
	if !ok1 || !ok2 {
		return
	}

	// Only act on Status changes to reduce noise
	if equality.Semantic.DeepEqual(oldE.Status, newE.Status) {
		return
	}

	state := newE.Status.Phase
	switch state {
	case shared.CompletedAS:
		t.depot.MarkArtifactCompleted(artifactKey(newE), shared.CompletedAS)
	case shared.FailedAS:
		t.depot.MarkArtifactCompleted(artifactKey(newE), shared.FailedAS)
	case shared.EnchantingAS:
		t.depot.UpdatePendingArtifact(artifactKey(newE), shared.EnchantingAS)
	case shared.RequeuedAS:
		t.depot.UpdatePendingArtifact(artifactKey(newE), shared.RequeuedAS)
	case shared.ScheduledAS:
	}

	t.logger.Info("enchantment update", slog.String("name", newE.Name), slog.String("state", state.String()))
}

func (t *EnchantmentTracker) onEnchantDelete(obj any) {
	ench, ok := obj.(*enchantv1.Enchantment)
	if !ok {
		if tomb, ok := obj.(cache2.DeletedFinalStateUnknown); ok {
			if cast, ok2 := tomb.Obj.(*enchantv1.Enchantment); ok2 {
				ench = cast
			}
		}
	}
	if ench == nil {
		return
	}

	state := ench.Status.Phase
	if state == shared.CompletedAS {
		t.depot.MarkArtifactCompleted(artifactKey(ench), shared.CompletedAS)
	} else if state == shared.FailedAS {
		t.depot.MarkArtifactCompleted(artifactKey(ench), shared.FailedAS)
	} else {
		// unexpected delete ?
		t.logger.Info("enchantment delete unexpected", slog.String("name", ench.Name), slog.String("last_state", state.String()))

		t.depot.MarkArtifactCompleted(artifactKey(ench), shared.FailedAS)
	}

	t.logger.Info("enchantment delete", slog.String("name", ench.Name), slog.String("last_state", state.String()))
}

func artifactKey(e *enchantv1.Enchantment) string {
	return string(e.GetUID())
}
