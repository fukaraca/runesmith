package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

func (p *DevicePlugin) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/allocations", p.handleAllocation)
	mux.HandleFunc("/healthz", p.handleHealthz)
	mux.HandleFunc("/readyz", p.handleReadyz)
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
	var ai AllocationInfo
	if err := json.NewDecoder(r.Body).Decode(&ai); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p.manager.MapAllocations(ai.PodUID, ai.PodName, ai.Namespace, ai.DeviceIDs, ai.Timestamp)
	w.WriteHeader(http.StatusCreated)
}

func (p *DevicePlugin) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func (p *DevicePlugin) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
