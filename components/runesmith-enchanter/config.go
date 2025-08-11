package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fukaraca/runesmith/shared"
	"github.com/spf13/viper"
)

type AppConfig struct {
	PodUID            string
	PodName           string
	Namespace         string
	TaskID            string
	ArtifactID        string
	DaemonServiceAddr string // includes http://serviceName and :port
	EnergyType        shared.Elemental
	DeviceIDs         []string
	DeviceCount       int
	EnchantmentCost   int
	HTTPPort          string
	SelfReport        bool
}

func readConfig() (*AppConfig, error) {
	viper.AutomaticEnv()

	viper.SetDefault("TASK_ID", "unknown") //
	viper.SetDefault("ARTIFACT_ID", "unknown")
	viper.SetDefault("HTTP_PORT", 8080)
	viper.SetDefault("ENCHANTMENT_COST", 10)
	viper.SetDefault("SELF_REPORT", true)

	cfg := &AppConfig{
		PodUID:            viper.GetString("POD_UID"),
		PodName:           viper.GetString("POD_NAME"),
		Namespace:         viper.GetString("POD_NAMESPACE"),
		TaskID:            viper.GetString("TASK_ID"),
		ArtifactID:        viper.GetString("ARTIFACT_ID"),
		DaemonServiceAddr: viper.GetString("DAEMON_SERVICE_ADDR"),
		EnergyType:        shared.Elemental(viper.GetString("MANA_ENERGY_TYPE")),
		DeviceIDs:         strings.Split(viper.GetString("MANA_DEVICE_IDS"), ","),
		DeviceCount:       viper.GetInt("MANA_COUNT"),
		EnchantmentCost:   viper.GetInt("ENCHANTMENT_COST"),
		HTTPPort:          viper.GetString("HTTP_PORT"),
		SelfReport:        viper.GetBool("SELF_REPORT"),
	}
	if len(cfg.DeviceIDs) == 0 || cfg.DeviceIDs[0] == "" {
		return nil, errors.New("device ids not found")
	}
	fmt.Printf("%+v\n", *cfg)
	return cfg, nil
}
