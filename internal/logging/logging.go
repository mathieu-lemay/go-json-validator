package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ConfigureLogger(opts ...LogConfig) error {
	config := ConfigureLoggerOptions{
		Level: zerolog.InfoLevel,
	}

	for _, opt := range opts {
		opt(&config)
	}

	zerolog.SetGlobalLevel(config.Level)

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Logger = logger

	return nil
}

type ConfigureLoggerOptions struct {
	Level zerolog.Level
}

type LogConfig func(*ConfigureLoggerOptions)

func WithLevel(l zerolog.Level) LogConfig {
	return func(c *ConfigureLoggerOptions) {
		c.Level = l
	}
}
