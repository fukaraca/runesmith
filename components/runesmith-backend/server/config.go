package server

import (
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/api/kubeapi"
	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/service"
	"github.com/gin-gonic/gin"
)

const V1 = "/api/v1"

type Server struct {
	Config  *config.Config
	engine  *gin.Engine
	router  *gin.RouterGroup
	Logger  *slog.Logger
	Service *service.Service
}

func NewServer(cfg *config.Config, engine *gin.Engine, logger *slog.Logger) (*Server, error) {
	v1 := engine.Group(V1)
	apiClient, err := kubeapi.NewInCluster(cfg.Metadata.Namespace) // use dedicated config after CRD
	if err != nil {
		return nil, err
	}
	svc := service.New(apiClient, cfg.Items, cfg.Plugin)

	return &Server{
		Config:  cfg,
		Logger:  logger,
		router:  v1,
		engine:  engine,
		Service: svc,
	}, nil
}
