package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	grpcServer *grpc.Server
	healthSrv  *health.Server
}

func NewHealthServer(grpcServer *grpc.Server) *HealthServer {
	hs := &HealthServer{
		grpcServer: grpcServer,
		healthSrv:  health.NewServer(),
	}
	hs.healthSrv.SetServingStatus(service_name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	return hs
}

func (h *HealthServer) Start() {
	grpc_health_v1.RegisterHealthServer(h.grpcServer, h.healthSrv)
	h.healthSrv.SetServingStatus(service_name, grpc_health_v1.HealthCheckResponse_SERVING)
}

func (h *HealthServer) Stop() {
	h.healthSrv.SetServingStatus(service_name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}
