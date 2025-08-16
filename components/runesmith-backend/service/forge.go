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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    shared.ScheduledAS,
	}

	job, err := s.kubeApi.CreateFireEnchantmentJob(ctx, art.ID, s.enchanter)
	if err != nil {
		return "", err
	}

	art.TaskID = string(job.UID)
	art.UpdatedAt = time.Now()
	s.depot.ScheduleNewArtifact(art)

	logger.Info("forge scheduled",
		"artifact_id", art.ID,
		"item_id", art.ItemID,
		"job_name", job.Name,
		"job_uid", string(job.UID),
	)
	return job.Name, nil
}
