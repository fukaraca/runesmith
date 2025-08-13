package server

import (
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/server/handlers"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Server, logger *slog.Logger, opts ...gin.OptionFunc) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	e := gin.New(opts...)
	e.Use(gin.Recovery(), LoggerMw(logger), ErrorHandlerMw(), CounterUIMw())
	return e
}

func (s *Server) bindRoutes() {
	s.engine.NoRoute(handlers.NoRoute404)

	r := handlers.NewRestHandler()
	s.engine.GET("/healthz", r.Healthz)
	s.router.GET("some", nil)
}
