package server

import (
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/gin-gonic/gin"
)

const V1 = "/api/v1"

type Server struct {
	Config *config.Config
	engine *gin.Engine
	router *gin.RouterGroup
	Logger *slog.Logger
}

func NewServer(cfg *config.Config, engine *gin.Engine, logger *slog.Logger) (*Server, error) {
	v1 := engine.Group(V1)

	return &Server{
		Config: cfg,
		Logger: logger,
		router: v1,
		engine: engine,
	}, nil
}
