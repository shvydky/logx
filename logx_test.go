package logx_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/shvydky/logx"
)

type testHandler struct {
	count int
}

type testTarget struct{}

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

func TestTypeLevelCanBeLowerThanDefaultLevel(t *testing.T) {
	var output bytes.Buffer

	logx.Init(&logx.Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			"github.com/shvydky/logx_test.testTarget": slog.LevelDebug,
		},
	}, logx.WithWriter(&output))

	logx.For(testTarget{}).Debug("type debug message")
	logx.For(t).Debug("default debug message")

	if !strings.Contains(output.String(), "type debug message") {
		t.Fatalf("type-level Debug override was filtered by the base handler: %q", output.String())
	}
	if strings.Contains(output.String(), "default debug message") {
		t.Fatalf("default Info level did not filter a Debug record: %q", output.String())
	}
}

func TestPackageLevelCanBeLowerThanDefaultLevelInPrettyMode(t *testing.T) {
	var output bytes.Buffer

	logx.Init(&logx.Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			"github.com/shvydky/logx_test": slog.LevelDebug,
		},
		Pretty: true,
	}, logx.WithWriter(&output))

	logx.For(testTarget{}).Debug("package debug message")

	if !strings.Contains(output.String(), "package debug message") {
		t.Fatalf("package-level Debug override was filtered by the base handler: %q", output.String())
	}
}
