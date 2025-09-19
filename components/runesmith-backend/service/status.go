package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fukaraca/runesmith/components/runesmith-backend/server/middlewares"
	"github.com/fukaraca/runesmith/shared"
)

const statusAPIPlugin = "v1/status"

func (s *Service) Status(ctx context.Context) ([]shared.NodeStatus, error) {
	s.StatusPoller.Ping()
	if v := s.StatusPoller.Latest(); v != nil {
		return v, nil
	}
	// as fallback get status directly
	return s.StatusGetter(ctx, middlewares.GetLoggerFromContext(ctx))
}

func (s *Service) StatusGetter(ctx context.Context, logger *slog.Logger) ([]shared.NodeStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	out := make([]shared.NodeStatus, len(s.plugin.Services))
	p := s.plugin

	for i := 0; i < len(p.Services); i++ {
		url := fmt.Sprintf("http://%s:%s/%s", p.Services[i], p.Port, statusAPIPlugin)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			logger.Error("StatusGetter.NewRequestWithContext failed", err)
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error("request to plugin failed", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			logger.Error("unexpected status code from plugin", slog.Int("status code", resp.StatusCode))
			resp.Body.Close()
			continue
		}
		var ns shared.NodeStatus
		if err := json.NewDecoder(resp.Body).Decode(&ns); err != nil {
			logger.Error("unexpected status code from plugin", slog.Int("status code", resp.StatusCode))
			resp.Body.Close()
			continue
		}
		out[i] = ns
		resp.Body.Close()
	}
	return out, nil
}
