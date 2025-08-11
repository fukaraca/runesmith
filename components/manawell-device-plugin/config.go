package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fukaraca/runesmith/shared"
	logg "github.com/fukaraca/runesmith/shared/log"
	"github.com/spf13/viper"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	service_name = "manawell-device-plugin"
)

func NewConfig() *Config {
	return &Config{}
}

var envs = []string{
	"mana.energyType",
	"node.name",
	"node.namespace",
	"node.daemonServiceName",
}

func (c *Config) Load(filename, path string) error {
	v := viper.New()
	v.SetConfigName(filename)
	v.AddConfigPath(path)
	v.SetConfigType("yml")

	v.SetDefault("kubelet.socketPath", v1beta1.DevicePluginPath+v1beta1.KubeletSocket)
	v.SetDefault("server.socketPath", v1beta1.DevicePluginPath+"manawell.sock")
	v.SetDefault("mana.maxMana", 100)
	v.SetDefault("node.namespace", "default")
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for _, env := range envs {
		if err := v.BindEnv(env); err != nil {
			return err
		}
	}

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(c)
	if err != nil {
		return err
	}
	c.Mana.ResourceName = fmt.Sprintf("manawell.io/%s", c.Mana.EnergyType)
	return nil
}

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Mana       ManaConfig       `mapstructure:"mana"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Kubelet    KubeletConfig    `mapstructure:"kubelet"`
	Log        logg.Config      `mapstructure:"log"`
	Watcher    WatcherConfig    `mapstructure:"watcher"`
	Node       NodeConfig       `mapstructure:"node"`
}

type ServerConfig struct {
	SocketPath string        `mapstructure:"socketPath"`
	Address    string        `mapstructure:"address"`
	Port       string        `mapstructure:"port"`
	Timeout    time.Duration `mapstructure:"timeout"`
}

type ManaConfig struct {
	MaxMana      int              `mapstructure:"maxMana"`
	EnergyType   shared.Elemental `mapstructure:"energyType"`
	ResourceName string           `mapstructure:"resourceName"`
}

type MonitoringConfig struct {
	MetricsPort    int           `mapstructure:"metricsPort"`
	UpdateInterval time.Duration `mapstructure:"updateInterval"`
}

type KubeletConfig struct {
	SocketPath      string        `mapstructure:"socketPath"`
	RetryAttempts   int           `mapstructure:"retryAttempts"`
	BackoffInterval time.Duration `mapstructure:"backoffInterval"`
}

type WatcherConfig struct {
	ResyncInterval      time.Duration `mapstructure:"resyncInterval"`
	SocketCheckInterval time.Duration `mapstructure:"socketCheckInterval"`
}

type NodeConfig struct {
	Name              string `mapstructure:"name"`
	Namespace         string `mapstructure:"namespace"`
	DaemonServiceName string `mapstructure:"daemonServiceName"`
}
