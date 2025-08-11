package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	logg "github.com/fukaraca/runesmith/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type DevicePlugin struct {
	logger       *slog.Logger
	config       *Config
	manager      *ManaGer
	grpcServer   *grpc.Server
	httpServer   *http.Server
	healthServer *HealthServer
	watcher      *PodWatcher
	stopCh       chan bool
}

func NewManaDevicePlugin(config *Config, manager *ManaGer) *DevicePlugin {
	return &DevicePlugin{
		logger:  logg.New(config.Log),
		config:  config,
		manager: manager,
		stopCh:  make(chan bool),
	}
}

func (p *DevicePlugin) Start(ctx context.Context) error {
	if err := p.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup previous socket: %w", err)
	}

	if err := p.serve(ctx); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	if err := p.registerWithKubelet(ctx); err != nil {
		return fmt.Errorf("failed to register with kubelet: %w", err)
	}

	var err error
	p.watcher, err = NewPodWatcher(p.config, p.manager, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create pod watcher: %w", err)
	}
	go p.watcher.Start(ctx)
	go p.watchKubeletRestart(ctx)

	go p.startHTTPServer()

	p.healthServer = NewHealthServer(p.grpcServer)
	go p.healthServer.Start()

	p.logger.Info("device plugin started successfully", slog.String("resource_name", p.config.Mana.ResourceName))
	return nil
}

func (p *DevicePlugin) Stop() {
	log.Println("Stopping device plugin...")

	close(p.stopCh)

	if p.grpcServer != nil {
		p.grpcServer.Stop()
	}

	if p.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		p.httpServer.Shutdown(ctx)
		cancel()
	}

	if p.healthServer != nil {
		p.healthServer.Stop()
	}

	if p.watcher != nil {
		p.watcher.Stop()
	}

	p.cleanup()
}

func (p *DevicePlugin) serve(ctx context.Context) error {
	sock, err := net.Listen("unix", p.config.Server.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", p.config.Server.SocketPath, err)
	}

	p.grpcServer = grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(p.grpcServer, p)

	go func() {
		if err := p.grpcServer.Serve(sock); err != nil {
			p.logger.Error("gRPC server error", err) // TODO errgroup or healthchecker
		}
	}()

	conn, err := grpc.NewClient("unix://"+p.config.Server.SocketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", addr, p.config.Server.Timeout)
		}))
	if err != nil {
		return fmt.Errorf("failed to dial device plugin socket: %w", err)
	}
	defer conn.Close()

	return nil
}

func (p *DevicePlugin) registerWithKubelet(ctx context.Context) error {
	conn, err := grpc.NewClient("unix://"+p.config.Kubelet.SocketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial kubelet socket: %w", err) // no need to retry just kill it
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)

	request := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     filepath.Base(p.config.Server.SocketPath),
		ResourceName: p.config.Mana.ResourceName, // no need to set any options
	}

	for attempt := 1; attempt < p.config.Kubelet.RetryAttempts+1; attempt++ {
		ctx, cancel := context.WithTimeout(ctx, p.config.Server.Timeout)
		_, err = client.Register(ctx, request)
		cancel()
		if err == nil {
			p.logger.Info("successfully registered with kubelet")
			return nil
		}

		p.logger.Warn("registration attempt failed", slog.Int("attempt", attempt), slog.Any("error", err))
		if attempt < p.config.Kubelet.RetryAttempts {
			time.Sleep(p.config.Kubelet.BackoffInterval)
		}
	}

	return fmt.Errorf("failed to register with kubelet after %d attempts: %w", p.config.Kubelet.RetryAttempts, err)
}

func (p *DevicePlugin) watchKubeletRestart(ctx context.Context) {
	ticker := time.NewTicker(p.watcher.watcher.SocketCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := os.Stat(p.config.Server.SocketPath); os.IsNotExist(err) { // kubelet restart deleted our socket, lets restart
				p.logger.Warn("kubelet restarted, restarting everything")
				p.Stop()
				if err := p.Start(ctx); err != nil {
					p.logger.Error("Failed to restart", err)
					p.Stop()
				} else {
					return // we are running another watcher already
				}
			}
		}
	}
}

func (p *DevicePlugin) cleanup() error {
	if _, err := os.Stat(p.config.Server.SocketPath); err == nil {
		// lets delete old sock
		if err := os.Remove(p.config.Server.SocketPath); err != nil {
			return fmt.Errorf("failed to remove socket file: %w", err)
		}
	}
	return nil
}

func (p *DevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired:                false,
		GetPreferredAllocationAvailable: false,
	}, nil
}

func (p *DevicePlugin) ListAndWatch(empty *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	devices := p.manager.GetAllDevices()
	if err := stream.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Minute * 5) // there is no meaning of healthiness on our mana droplets anyway
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return nil
		case <-ticker.C:
			if err := stream.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
				return err
			}
		}
	}
}

func (p *DevicePlugin) Allocate(ctx context.Context, req *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	response := &pluginapi.AllocateResponse{
		ContainerResponses: make([]*pluginapi.ContainerAllocateResponse, len(req.ContainerRequests)),
	}

	for i, containerReq := range req.ContainerRequests {
		count := len(containerReq.DevicesIDs)
		allocatedIDs, err := p.manager.AllocateDevices(count)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate devices: %w", err)
		}

		containerResponse := &pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"MANA_ENERGY_TYPE": p.config.Mana.EnergyType.String(),
				"MANA_DEVICE_IDS":  strings.Join(allocatedIDs, ","),
				"MANA_COUNT":       fmt.Sprintf("%d", count),
			},
		}

		response.ContainerResponses[i] = containerResponse
		p.logger.Info(fmt.Sprintf("allocated %d mana devices: %v", count, allocatedIDs))
	}

	return response, nil
}

func (p *DevicePlugin) GetPreferredAllocation(ctx context.Context, req *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (p *DevicePlugin) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *DevicePlugin) mustEmbedUnimplementedDevicePluginServer() {}
