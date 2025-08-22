package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fukaraca/runesmith/shared"
)

func (p *DevicePlugin) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/allocations", p.handleAllocation)
	mux.HandleFunc("/healthz", p.handleHealthz)
	mux.HandleFunc("/readyz", p.handleReadyz)
	mux.HandleFunc("/v1/status", p.handleStatus)
	srv := &http.Server{
		Addr:              net.JoinHostPort(p.config.Server.Address, p.config.Server.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("device plugin http server failed: %v", err)
		}
	}()
	p.httpServer = srv
}

func (p *DevicePlugin) handleAllocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var ai shared.AllocationInfo
	if err := json.NewDecoder(r.Body).Decode(&ai); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p.manager.MapAllocations(ai.PodUID, ai.PodName, ai.Namespace, ai.DeviceIDs, ai.Timestamp)
	w.WriteHeader(http.StatusCreated)
}

func (p *DevicePlugin) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (p *DevicePlugin) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (p *DevicePlugin) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := shared.NodeStatus{
		Name:        p.config.Mana.ResourceName,
		Available:   p.manager.GetAvailableMana(),
		Allocated:   p.manager.GetAllocatedMana(),
		RunningJobs: len(p.manager.allocations),
	}
	if _, err := os.Stat(p.config.Kubelet.SocketPath); err == nil {
		status.Healthy = true
	} else {
		p.logger.Warn("kubelet socket check failed", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
