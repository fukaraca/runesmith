package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	logg "github.com/fukaraca/runesmith/shared/log"
)

func run() error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}

	logger := logg.New(logg.Config{
		Level:     "info",
		AddSource: true,
	})
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL) // in case of preemption
	defer cancel()
	if cfg.SelfReport {
		if err := postAllocation(ctx, logger, cfg); err != nil {
			logger.Error("self-report to daemon failed; exiting", err)
			return err
		}
	} else {
		// for off-cluster testability
	}

	httpSrv, err := startHTTP(cfg.HTTPPort)
	if err != nil {
		return err
	}

	m := cfg.DeviceCount
	totalSeconds, elapsed := cfg.DeviceCount*cfg.EnchantmentCost, 0
	logger.Info("begin enchantment", slog.Any("energy", cfg.EnergyType), slog.Int("device count", m), "duration_seconds", totalSeconds)

	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for m > 0 {
		select {
		case <-ctx.Done():
			logger.Warn("received shutdown signal; aborting job", slog.Int("elapsed_seconds", elapsed))
			m = 0
			break
		case <-ticker.C:
			elapsed++
			if elapsed%cfg.EnchantmentCost == 0 {
				m--
				logger.Info("enchantment progress",
					slog.Int("percent", (elapsed*100)/totalSeconds),
					slog.Int("elapsed_s", elapsed),
					slog.Int("remaining_s", totalSeconds-elapsed))
			}
		}
	}

	httpSrv.Shutdown(ctx)

	logger.Info("enchantment finished", slog.String("total_elapsed", time.Since(start).String()))
	return nil
}
