package service

import (
	"log/slog"
	"sync/atomic"

	"github.com/fukaraca/runesmith/components/runesmith-backend/api/kubeapi"
	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"
	"github.com/fukaraca/runesmith/shared"
)

type Service struct {
	depot     *artifactory.Artifactory
	Items     []shared.MagicalItem
	counter   atomic.Uint64
	kubeApi   *kubeapi.Client
	plugin    config.Plugin
	enchanter config.Enchanter
	Tracker   *kubeapi.EnchantmentTracker
}

func (s *Service) nextID() int {
	return int(s.counter.Add(1))
}

func New(api *kubeapi.Client, items []shared.MagicalItem, plugin config.Plugin, enchanter config.Enchanter, meta *config.Meta, logger *slog.Logger) (*Service, error) {
	art := artifactory.NewArtifactory()
	tracker, err := kubeapi.NewEnchantmentTracker(api, meta, logger, art)
	if err != nil {
		return nil, err
	}
	return &Service{
		Items:     items,
		depot:     art,
		kubeApi:   api,
		plugin:    plugin,
		enchanter: enchanter,
		Tracker:   tracker,
	}, nil
}
