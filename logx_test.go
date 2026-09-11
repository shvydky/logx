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
	handler := &testHandler{}
	logx.Init(config, logx.WithHandler(handler))

	logx.For(nil).Info("This should not panic")
	logx.For(t).Info("Should log with type info", slog.String("test", "logx_test"))
	logx.For(t).Debug("Should log with type debug", slog.String("test", "logx_test"))
	logx.For(config).Debug("Should be ignored")

	if handler.count != 3 {
		t.Errorf("expected 3 log entries handled, got %d", handler.count)
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

func TestDefaultLoggerKeepsGlobalLevelWithLowerOverride(t *testing.T) {
	for _, pretty := range []bool{false, true} {
		t.Run(map[bool]string{false: "JSON", true: "pretty"}[pretty], func(t *testing.T) {
			var output bytes.Buffer

			logx.Init(&logx.Config{
				DefaultLevel: slog.LevelInfo,
				Levels: map[string]slog.Level{
					"github.com/shvydky/logx_test.testTarget": slog.LevelDebug,
				},
				Pretty: pretty,
			}, logx.WithWriter(&output))

			slog.Default().Debug("default debug message")
			logx.For(nil).Debug("nil target debug message")
			logx.For(testTarget{}).Debug("overridden debug message")

			got := output.String()
			if strings.Contains(got, "default debug message") {
				t.Fatalf("default logger bypassed the global level: %q", got)
			}
			if strings.Contains(got, "nil target debug message") {
				t.Fatalf("nil target bypassed the global level: %q", got)
			}
			if !strings.Contains(got, "overridden debug message") {
				t.Fatalf("type-level override was filtered: %q", got)
			}
		})
	}
}

func TestInitTakesConfigurationSnapshot(t *testing.T) {
	var output bytes.Buffer
	config := &logx.Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			"github.com/shvydky/logx_test.testTarget": slog.LevelDebug,
		},
	}
	logx.Init(config, logx.WithWriter(&output))

	config.DefaultLevel = slog.LevelError
	config.Levels["github.com/shvydky/logx_test.testTarget"] = slog.LevelError

	logx.For(t).Info("snapshot info message")
	logx.For(testTarget{}).Debug("snapshot debug message")

	got := output.String()
	if !strings.Contains(got, "snapshot info message") {
		t.Fatalf("mutating DefaultLevel changed the active configuration: %q", got)
	}
	if !strings.Contains(got, "snapshot debug message") {
		t.Fatalf("mutating Levels changed the active configuration: %q", got)
	}
}

func TestForUsesCurrentDefaultHandler(t *testing.T) {
	first := &testHandler{}
	second := &testHandler{}

	logx.Init(&logx.Config{DefaultLevel: slog.LevelInfo}, logx.WithHandler(first))
	slog.SetDefault(slog.New(second))

	logx.For(testTarget{}).Info("current handler message")

	if first.count != 0 {
		t.Fatalf("original handler received %d records, want 0", first.count)
	}
	if second.count != 1 {
		t.Fatalf("current default handler received %d records, want 1", second.count)
	}
}
