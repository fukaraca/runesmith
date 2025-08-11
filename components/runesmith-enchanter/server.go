package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fukaraca/runesmith/shared"
)

func postAllocation(ctx context.Context, logger *slog.Logger, cfg *AppConfig) error {
	body := shared.AllocationInfo{
		PodUID:    cfg.PodUID,
		PodName:   cfg.PodName,
		Namespace: cfg.Namespace,
		DeviceIDs: cfg.DeviceIDs,
		Timestamp: time.Now().Unix(),
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := cfg.DaemonServiceAddr + "/v1/allocations"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("allocation report failed status=%d", resp.StatusCode)
	}
	logger.Info("reported allocation to daemon", "url", url, "deviceIDs", strings.Join(cfg.DeviceIDs, ","))

	return nil
}

func startHTTP(port string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", handleReadyz)
	srv := &http.Server{
		Addr:              net.JoinHostPort("", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("device plugin http server failed: %v", err)
		}
	}()

	return srv, nil
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
