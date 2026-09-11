package logx

import (
	"log/slog"
	"os"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/lmittmann/tint"
)

var config atomic.Pointer[levelConfig]

const (
	attrPkg  = "pkg"
	attrType = "type"
)

func Init(cfg *Config, opts ...Options) *slog.Logger {
	if cfg == nil {
		cfg = &Config{}
	}

	for _, opt := range opts {
		opt(cfg)
	}

	lc := newLevelConfig(cfg)

	var h slog.Handler
	if cfg.handler != nil {
		h = cfg.handler
	} else {
		w := cfg.writer
		if w == nil {
			w = os.Stdout
		}

		if lc.src.Pretty {
			h = tint.NewHandler(w, &tint.Options{Level: lc.minLevel(), TimeFormat: time.RFC3339})
		} else {
			h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lc.minLevel()})
		}
	}

	lc.handler = h
	config.Store(lc)
	slog.SetDefault(slog.New(newLevelHandler(h, cfg.DefaultLevel)))
	return slog.Default()
}

func For(target any) *slog.Logger {
	if target == nil {
		return slog.Default()
	}

	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	pkg := t.PkgPath()
	typ := t.Name()
	full := typ

	if pkg != "" && typ != "" {
		full = pkg + "." + typ
	}

	cfg := config.Load()
	level := slog.LevelInfo
	next := slog.Default().Handler()
	if cfg != nil {
		level = cfg.src.DefaultLevel
		next = cfg.handler
		if l, ok := cfg.byType[full]; ok {
			level = l
		} else if pkg != "" {
			if l, ok := cfg.byPackage[pkg]; ok {
				level = l
			}
		}
	}
	logger := slog.New(newLevelHandler(next, level))

	if pkg != "" {
		logger = logger.With(slog.String(attrPkg, pkg))
	}
	if typ != "" {
		logger = logger.With(slog.String(attrType, typ))
	}

	return logger
}
