package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version    string = "dev"
	configFile string
)

func RootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "manawell-device-plugin",
		Short:   "Kubernetes Device Plugin for Mana resources",
		Long:    "A Kubernetes Device Plugin that manages magical(virtual) Mana resources for Runesmith project",
		RunE:    func(cmd *cobra.Command, args []string) error { return runPlugin() },
		Version: Version,
	}
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.example.yaml", "Path to configuration file")
	rootCmd.AddCommand(loadConfig())
	return rootCmd
}

func loadConfig() *cobra.Command {
	return &cobra.Command{
		Use:   "load-config",
		Short: "load config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := NewConfig()
			err := cfg.Load(configFile, "./configs")
			if err != nil {
				return err
			}
			b, _ := json.MarshalIndent(*cfg, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
}

func runPlugin() error {
	config := NewConfig()
	err := config.Load(configFile, "./configs")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	manager := NewManaGer(config.Mana)

	//	metricsServer := NewMetricsServer(config.Monitoring.MetricsPort, manager)
	//	go metricsServer.Start()

	plugin := NewManaDevicePlugin(config, manager)

	go func() {
		if err := plugin.Start(ctx); err != nil {
			log.Fatalf("Failed to start device plugin: %v", err)
		}
	}()

	<-signalChan
	plugin.logger.Info("Received shutdown signal, gracefully shutting down...")

	cancel()
	plugin.Stop()
	time.Sleep(2 * time.Second)
	return err
}

func main() {
	if err := RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
