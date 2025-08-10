package logg

import (
	"log/slog"
	"os"
)

type LogLevel string
type Config struct {
	Level     LogLevel `mapstructure:"level"`
	AddSource bool     `mapstructure:"addSource"`
}

func (l LogLevel) Int() slog.Level {
	switch l {
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

func New(cfg Config) *slog.Logger {
	opt := slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level.Int(),
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opt))
	slog.SetDefault(logger)
	return logger
}
