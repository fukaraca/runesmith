package server

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	logg "github.com/fukaraca/runesmith/shared/log"
	gb "github.com/fukaraca/skypiea/pkg/guest_book"
)

func Start(cfg *config.Config) error {
	logger := logg.New(cfg.Log)
	router := NewRouter(cfg.Server, logger)
	router.SetTrustedProxies(nil)
	server, err := NewServer(cfg, router, logger)
	if err != nil {
		return err
	}
	server.bindRoutes()
	gb.GuestBook = gb.New() // do we really get benefit

	logger.Info("Server started")
	httpServer := &http.Server{
		Addr:              net.JoinHostPort(cfg.Server.Address, cfg.Server.Port),
		Handler:           server.engine,
		ReadHeaderTimeout: time.Second * 5,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Warn("received interrupt signal")
		if errInner := httpServer.Close(); errInner != nil {
			log.Fatal("Server Close:", errInner)
		}
	}()

	go func() {
		if errInner := server.Service.Tracker.Start(context.Background()); errInner != nil {
			log.Fatalf("tracker failed %v", errInner)
		}
	}()

	if err = httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Warn("Server closed under request")
		} else {
			log.Fatal("Server closed unexpectedly")
		}
	}
	server.Service.Tracker.Stop()
	logger.Info("Server shutting down")
	return nil
}
