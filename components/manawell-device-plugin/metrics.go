package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO not sure if i am gonna use this but lets keep it

type MetricsServer struct {
	port           int
	updateInterval time.Duration
	manager        *ManaGer
	manaGauge      prometheus.Gauge
	allocGauge     prometheus.Gauge
	registry       *prometheus.Registry
	server         *http.Server
}

func NewMetricsServer(config MonitoringConfig, manager *ManaGer) *MetricsServer {
	registry := prometheus.NewRegistry()

	manaGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "manawell_available_mana_total",
		Help: "Total available mana in the pool",
	})

	allocGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "manawell_allocated_mana_total",
		Help: "Total allocated mana",
	})

	registry.MustRegister(manaGauge)
	registry.MustRegister(allocGauge)

	return &MetricsServer{
		port:       config.MetricsPort,
		manager:    manager,
		manaGauge:  manaGauge,
		allocGauge: allocGauge,
		registry:   registry,
	}
}

func (m *MetricsServer) Start() error {
	go m.updateMetrics()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	log.Printf("Metrics server starting on port %d", m.port)
	return m.server.ListenAndServe()
}

func (m *MetricsServer) updateMetrics() {
	ticker := time.NewTicker(m.updateInterval)
	defer ticker.Stop()

	for range ticker.C {
		available := m.manager.GetAvailableMana()
		allocated := m.manager.GetAllocatedMana()

		m.manaGauge.Set(float64(available))
		m.allocGauge.Set(float64(allocated))
	}
}

func (m *MetricsServer) Stop() {
	if m.server != nil {
		m.server.Close()
	}
}
