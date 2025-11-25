package logx_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/shvydky/logx"
)

type testHandler struct {
	count int
}

// Enabled implements slog.Handler.
func (t *testHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (t *testHandler) Handle(context.Context, slog.Record) error {
	t.count++
	return nil
}

// WithAttrs implements slog.Handler.
func (t *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return t
}

// WithGroup implements slog.Handler.
func (t *testHandler) WithGroup(name string) slog.Handler {
	return t
}

func TestFor(t *testing.T) {
	config := &logx.Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			"testing.T": slog.LevelDebug,
		},
		Pretty: true,
	}
	logx.Init(config, logx.WithHandler(&testHandler{}))

	logx.For(nil).Info("This should not panic")
	logx.For(t).Info("Should log with type info", slog.String("test", "logx_test"))
	logx.For(t).Debug("Should log with type debug", slog.String("test", "logx_test"))
	logx.For(config).Debug("Should be ignored")

	if th, ok := slog.Default().Handler().(*testHandler); ok {
		if th.count != 3 {
			t.Errorf("expected 3 log entries handled, got %d", th.count)
		}
	}
}
