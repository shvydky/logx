package logx

import (
	"log/slog"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lmittmann/tint"
)

var config atomic.Pointer[levelConfig]
var stateMu sync.RWMutex

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

		if cfg.Pretty {
			h = tint.NewTextHandler(w, &tint.Options{Level: lc.minLevel(), TimeFormat: time.RFC3339})
		} else {
			h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lc.minLevel()})
		}
	}

	logger := slog.New(newLevelHandler(h, lc.defaultLevel))
	stateMu.Lock()
	config.Store(lc)
	slog.SetDefault(logger)
	stateMu.Unlock()
	return logger
}

func For(target any) *slog.Logger {
	if target == nil {
		return slog.Default()
	}

	stateMu.RLock()
	defer stateMu.RUnlock()

	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Pointer {
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
		level = cfg.levelFor(pkg, full)
	}
	logger := slog.New(withLevel(next, level))

	if pkg != "" {
		logger = logger.With(slog.String(attrPkg, pkg))
	}
	if typ != "" {
		logger = logger.With(slog.String(attrType, typ))
	}

	return logger
}
