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
	for _, opt := range opts {
		opt(cfg)
	}

	lc := newLevelConfig(cfg)
	config.Store(lc)

	if cfg.handler != nil {
		slog.SetDefault(slog.New(cfg.handler))
		return slog.Default()
	}

	w := cfg.writer
	if w == nil {
		w = os.Stdout
	}

	var h slog.Handler
	if lc.src.Pretty {
		h = tint.NewHandler(w, &tint.Options{Level: slog.LevelDebug, TimeFormat: time.RFC3339})
	} else {
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: cfg.DefaultLevel})
	}
	slog.SetDefault(slog.New(h))
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
	level := cfg.src.DefaultLevel
	if l, ok := cfg.byType[full]; ok {
		level = l
	} else if pkg != "" {
		if l, ok := cfg.byPackage[pkg]; ok {
			level = l
		}
	}
	logger := slog.New(newLevelHandler(slog.Default().Handler(), level))

	if pkg != "" {
		logger = logger.With(slog.String(attrPkg, pkg))
	}
	if typ != "" {
		logger = logger.With(slog.String(attrType, typ))
	}

	return logger
}
