package service

import (
	"sync/atomic"

	"github.com/fukaraca/runesmith/components/runesmith-backend/api/kubeapi"
	"github.com/fukaraca/runesmith/shared"
)

type Service struct {
	depot   *artifactory
	Items   []shared.MagicalItem
	counter atomic.Uint64
	kubeApi *kubeapi.Client
}

func (s *Service) nextID() int {
	return int(s.counter.Add(1))
}

func New(api *kubeapi.Client, items []shared.MagicalItem) *Service {
	return &Service{
		Items:   items,
		depot:   newArtifactory(),
		kubeApi: api,
	}
}
