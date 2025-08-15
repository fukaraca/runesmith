package handlers

import (
	"context"

	"github.com/fukaraca/runesmith/components/runesmith-backend/service"
	"github.com/fukaraca/runesmith/shared"
)

type ItemsService interface {
	AllItems() []shared.MagicalItem
	Forge(ctx context.Context) (string, error)
	GetArtifacts(completed bool) []service.Artifact
	Status()
}

type Rest struct {
	svc ItemsService
}

func NewRestHandler(svc *service.Service) *Rest {
	return &Rest{svc: svc}
}
