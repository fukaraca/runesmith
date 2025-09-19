package nodes

import (
	"context"
	"sync"
	"time"

	"github.com/fukaraca/runesmith/components/runesmith-backend/server/middlewares"
	"github.com/fukaraca/runesmith/shared"
)

// StatusPoller is helper that gets node statuses periodically but goes idle if no visitor to keep overhead minimal
type StatusPoller struct {
	mu          sync.Mutex
	running     bool
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	interval    time.Duration
	idleTimeout time.Duration
	rttTimeout  time.Duration

	lastSeen time.Time

	latestMu sync.RWMutex
	latest   []shared.NodeStatus

	getter func(ctx context.Context) ([]shared.NodeStatus, error)
}

func NewStatusPoller(getter func(ctx context.Context) ([]shared.NodeStatus, error), interval, idleTimeout time.Duration) *StatusPoller {
	return &StatusPoller{
		interval:    interval,
		rttTimeout:  time.Second * 5,
		idleTimeout: idleTimeout,
		getter:      getter,
	}
}

// Ping by each HTTP request to mark activity and ensure poller is running
func (p *StatusPoller) Ping() {
	p.lastSeen = time.Now()
	p.startIfNeeded()
}

func (p *StatusPoller) startIfNeeded() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		return
	}
	p.running = true
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.run(ctx)
	}()
}

func (p *StatusPoller) run(ctx context.Context) {
	// initial fetch so first visitor don't wait
	p.fetchOnce(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// if idle, stop
			if time.Since(p.lastSeen) > p.idleTimeout {
				// stop
				p.mu.Lock()
				if p.cancel != nil {
					p.cancel()
				}
				p.running = false
				p.mu.Unlock()
				return
			}
			p.fetchOnce(ctx)
		}
	}
}

func (p *StatusPoller) fetchOnce(ctx context.Context) {
	// short timeout per round
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := p.getter(cctx)
	if err != nil {
		middlewares.GetLoggerFromContext(ctx).Error("status poller fetch error:", err)
		return
	}
	p.latestMu.Lock()
	p.latest = out
	p.latestMu.Unlock()
}

func (p *StatusPoller) Latest() []shared.NodeStatus {
	p.latestMu.RLock()
	defer p.latestMu.RUnlock()
	if p.latest == nil {
		return nil
	}
	out := make([]shared.NodeStatus, len(p.latest))
	copy(out, p.latest)
	return out
}

func (p *StatusPoller) Stop() {
	p.mu.Lock()
	if p.cancel != nil {
		p.cancel()
	}
	p.running = false
	p.mu.Unlock()
	p.wg.Wait()
}
