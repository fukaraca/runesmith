package config

import (
	"strings"
	"time"

	"github.com/fukaraca/runesmith/shared"
	logg "github.com/fukaraca/runesmith/shared/log"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceMode string
	Server      Server               `mapstructure:"server"`
	Log         logg.Config          `mapstructure:"log"`
	Items       []shared.MagicalItem `mapstructure:"magicalItems"`
}

type Server struct {
	Address               string        `mapstructure:"address"`
	Port                  string        `mapstructure:"port"`
	MaxBodySizeMB         int           `mapstructure:"maxBodySizeMB"`
	GinMode               string        `mapstructure:"ginMode"`
	SessionTimeout        time.Duration `mapstructure:"sessionTimeout"`
	DefaultRequestTimeout time.Duration `mapstructure:"defaultRequestTimeout"`
	Version               string
}

func NewConfig() *Config {
	return &Config{}
}

var envs = []string{}

func (c *Config) Load(filename, path string) error {
	v := viper.New()
	v.SetConfigName(filename)
	v.AddConfigPath(path)
	v.SetConfigType("yml")

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

	return nil
}
