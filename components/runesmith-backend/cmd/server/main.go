package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/fukaraca/runesmith/components/runesmith-backend/config"
	"github.com/fukaraca/runesmith/components/runesmith-backend/server"
	logg "github.com/fukaraca/runesmith/shared/log"
	"github.com/spf13/cobra"
)

var (
	Version    string = "dev"
	configName string
)

func main() {
	if err := RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func RootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "runesmith-backend",
		Short: "Runesmith backend server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initialize()
		},
		Version: Version,
	}
	rootCmd.PersistentFlags().StringVar(&configName, "config", "config.example.yml", "config file name in configs folder")

	rootCmd.AddCommand(loadConfig())
	return rootCmd
}

func initialize() error {
	cfg := config.NewConfig()
	err := cfg.Load(configName, "./configs")
	if err != nil {
		return err
	}
	cfg.Server.Version = Version

	logg.New(cfg.Log).Info("server initialized", slog.Any("version", cfg.Server.Version))
	return server.Start(cfg)
}

func loadConfig() *cobra.Command {
	return &cobra.Command{
		Use:   "load-config",
		Short: "load config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.NewConfig()
			err := cfg.Load(configName, "./configs")
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", *cfg)
			return nil
		},
	}
}
