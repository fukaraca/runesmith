package artifactory

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fukaraca/runesmith/shared"
)

// Artifact is produce order of the item.
type Artifact struct {
	ID        int
	ItemID    int
	ItemName  string
	TaskID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    shared.EnchantmentPhase
}

type Artifactory struct {
	mu      sync.RWMutex
	pending atomic.Value // stores []Artifact
	done    atomic.Value // stores []Artifact
}

func NewArtifactory() *Artifactory {
	a := &Artifactory{}
	a.pending.Store([]Artifact{})
	a.done.Store([]Artifact{})
	return a
}

func (a *Artifactory) ScheduleNewArtifact(art *Artifact) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cur := a.pending.Load().([]Artifact)

	next := make([]Artifact, len(cur)+1)
	copy(next, cur)
	next[len(cur)] = *art
	a.pending.Store(next)
}

func (a *Artifactory) UpdatePendingArtifact(id string, status shared.EnchantmentPhase) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pending := a.pending.Load().([]Artifact)
	for i := range pending {
		if pending[i].TaskID == id {
			pending[i].Status = status
			pending[i].UpdatedAt = time.Now()
		}
	}
	a.pending.Store(pending)
}

func (a *Artifactory) MarkArtifactCompleted(id string, status shared.EnchantmentPhase) {
	a.mu.Lock()
	defer a.mu.Unlock()

	curP := a.pending.Load().([]Artifact)
	newP := make([]Artifact, 0, len(curP))
	var moved *Artifact
	for i := range curP {
		if curP[i].TaskID == id {
			tmp := curP[i] // value copy
			moved = &tmp
			continue
		}
		newP = append(newP, curP[i])
	}
	a.pending.Store(newP)

	if moved != nil {
		moved.Status = status
		moved.UpdatedAt = time.Now()
		curD := a.done.Load().([]Artifact)
		newD := make([]Artifact, len(curD)+1)
		copy(newD, curD)
		newD[len(curD)] = *moved
		a.done.Store(newD)
	}
}

func (a *Artifactory) Done() []Artifact {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.done.Load().([]Artifact)
}

func (a *Artifactory) Pending() []Artifact {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.pending.Load().([]Artifact)
}
