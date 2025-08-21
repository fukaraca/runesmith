package service

import (
	"context"
	"time"

	"github.com/fukaraca/runesmith/components/runesmith-backend/server/middlewares"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"
	"github.com/fukaraca/runesmith/shared"
)

func (s *Service) Forge(ctx context.Context) (string, error) {
	logger := middlewares.GetLoggerFromContext(ctx)
	id := s.nextID()
	item := s.randomItem()
	art := &artifactory.Artifact{
		ID:        id,
		ItemID:    item.ID,
		ItemName:  item.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    shared.ScheduledAS,
	}

	enchantment, err := s.kubeApi.CreateEnchantment(ctx, art, s.enchanter, item)
	if err != nil {
		return "", err
	}

	art.TaskID = string(enchantment.GetUID())
	art.UpdatedAt = time.Now()
	s.depot.ScheduleNewArtifact(art)

	logger.Info("forge scheduled",
		"artifact_id", art.ID,
		"item_id", art.ItemID,
		"enchantment_name", enchantment.GetName(),
		"enchantment_uid", string(enchantment.GetUID()),
	)
	return enchantment.GetName(), nil
}
