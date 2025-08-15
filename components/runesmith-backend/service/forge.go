package service

import (
	"context"
	"time"

	"github.com/fukaraca/runesmith/components/runesmith-backend/server/middlewares"
	"github.com/fukaraca/runesmith/shared"
	"github.com/google/uuid"
)

func (s *Service) Forge(ctx context.Context) (string, error) {
	logger := middlewares.GetLoggerFromContext(ctx)
	id := s.nextID()
	item := s.randomItem()
	art := &Artifact{
		ID:        id,
		ItemID:    item.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    shared.ScheduledAS,
	}

	job, err := s.kubeApi.CreateFireEnchantmentJob(ctx, art.ID, "ghcr.io/fukaraca/runesmith-enchanter:1.0.8", 20)
	if err != nil {
		return "", err
	}

	if uid, parseErr := uuid.Parse(string(job.UID)); parseErr == nil {
		art.TaskID = uid
		art.UpdatedAt = time.Now()
	} else {
		logger.Warn("unable to parse job UID into UUID", err) //???
	}
	s.depot.scheduleNewArtifact(art)

	logger.Info("forge scheduled",
		"artifact_id", art.ID,
		"item_id", art.ItemID,
		"job_name", job.Name,
		"job_uid", string(job.UID),
	)
	return job.Name, nil
}
