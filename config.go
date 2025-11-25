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
	src       *Config
	byPackage map[string]slog.Level
	byType    map[string]slog.Level
}

func newLevelConfig(cfg *Config) *levelConfig {
	lc := &levelConfig{
		src:       cfg,
		byPackage: make(map[string]slog.Level),
		byType:    make(map[string]slog.Level),
	}

	for key, lvl := range cfg.Levels {
		isType, normalized := classifyLevelKey(key)

		if isType {
			lc.byType[normalized] = lvl
		} else {
			lc.byPackage[normalized] = lvl
		}
	}

	return lc
}

func classifyLevelKey(key string) (bool, string) {
	k := strings.TrimSpace(key)
	if k == "" {
		return false, k
	}

	lastSlash := strings.LastIndex(k, "/")
	lastDot := strings.LastIndex(k, ".")

	return lastDot > lastSlash, k
}
