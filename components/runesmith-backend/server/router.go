package server

import (
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/server/handlers"
	"github.com/fukaraca/runesmith/components/runesmith-backend/server/middlewares"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Server, logger *slog.Logger, opts ...gin.OptionFunc) *gin.Engine {
	gin.SetMode(cfg.GinMode)
	e := gin.New(opts...)
	e.Use(gin.Recovery(), middlewares.LoggerMw(logger), middlewares.ErrorHandlerMw(), middlewares.CounterUIMw())
	return e
}

func (s *Server) bindRoutes() {
	s.engine.NoRoute(handlers.NoRoute404)

	r := handlers.NewRestHandler(s.Service)
	s.engine.GET("/healthz", r.Healthz)
	s.engine.GET("/readyz", r.Readyz)

	s.router.GET("/items", r.GetItemsList)
	s.router.POST("/forge", middlewares.RateLimiterMw(), r.Forge)
	s.router.GET("/artifacts", r.Artifacts)

	s.router.GET("/status", r.Status)
}
