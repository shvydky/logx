package logx

import (
	"io"
	"log/slog"
	"strings"
)

// Config holds the configuration for the logx logger.
type Config struct {
	// handler is the default slog.Handler to use. If nil, a JSON handler is used that writes to Writer.
	handler slog.Handler
	// writer is the destination for the log output. If nil, os.Stdout is used.
	writer io.Writer
	// DefaultLevel is the default log level.
	// It is used for packages and types not listed in Levels.
	DefaultLevel slog.Level
	// Levels is a map of package or type names to log levels.
	// Package names are in the form "github.com/user/project/pkg".
	// Type names are in the form "github.com/user/project/pkg.Type".
	// Type-specific levels take precedence over package-specific levels.
	Levels map[string]slog.Level
	// Pretty enables pretty printing of log entries
	Pretty bool
}

type Options func(*Config)

func WithWriter(w io.Writer) Options {
	return func(cfg *Config) {
		cfg.writer = w
	}
}

func WithHandler(h slog.Handler) Options {
	return func(cfg *Config) {
		cfg.handler = h
	}
}

type levelConfig struct {
	defaultLevel slog.Level
	levels       map[string]slog.Level
	handler      slog.Handler
}

func newLevelConfig(cfg *Config) *levelConfig {
	lc := &levelConfig{
		defaultLevel: cfg.DefaultLevel,
		levels:       make(map[string]slog.Level),
	}

	for key, lvl := range cfg.Levels {
		lc.levels[strings.TrimSpace(key)] = lvl
	}

	return lc
}

// minLevel returns the lowest level that can be enabled by the configuration.
// Built-in handlers must use this level so that their own filtering does not
// discard records allowed by a package- or type-specific override.
func (lc *levelConfig) minLevel() slog.Level {
	level := lc.defaultLevel

	for _, candidate := range lc.levels {
		if candidate < level {
			level = candidate
		}
	}

	return level
}

func (lc *levelConfig) levelFor(pkg, full string) slog.Level {
	if level, ok := lc.levels[full]; ok {
		return level
	}
	if level, ok := lc.levels[pkg]; ok {
		return level
	}
	return lc.defaultLevel
}
