package service

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fukaraca/runesmith/shared"
	"github.com/google/uuid"
)

// Artifact is produce order of the item.
type Artifact struct {
	ID        int
	ItemID    int
	TaskID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    shared.ArtifactStatus
}

type artifactory struct {
	mu      sync.Mutex
	pending atomic.Value // stores []Artifact
	done    atomic.Value // stores []Artifact
}

func newArtifactory() *artifactory {
	a := &artifactory{}
	a.pending.Store([]Artifact{})
	a.done.Store([]Artifact{})
	return a
}

func (a *artifactory) scheduleNewArtifact(art *Artifact) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cur := a.pending.Load().([]Artifact)

	next := make([]Artifact, len(cur)+1)
	copy(next, cur)
	next[len(cur)] = *art
	a.pending.Store(next)
}

func (a *artifactory) markArtifactCompleted(id int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	curP := a.pending.Load().([]Artifact)
	newP := make([]Artifact, 0, len(curP))
	var moved *Artifact
	for i := range curP {
		if curP[i].ID == id {
			tmp := curP[i] // value copy
			moved = &tmp
			continue
		}
		newP = append(newP, curP[i])
	}
	a.pending.Store(newP)

	if moved != nil {
		curD := a.done.Load().([]Artifact)
		newD := make([]Artifact, len(curD)+1)
		copy(newD, curD)
		newD[len(curD)] = *moved
		a.done.Store(newD)
	}
}

func (s *Service) GetArtifacts(completed bool) []Artifact {
	if completed {
		return s.depot.done.Load().([]Artifact)
	}
	return s.depot.pending.Load().([]Artifact)
}
