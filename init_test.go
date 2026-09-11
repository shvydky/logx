package logx

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

type fallbackTarget struct{}

func preserveGlobals(t *testing.T) {
	t.Helper()

	previousConfig := config.Load()
	previousLogger := slog.Default()
	t.Cleanup(func() {
		config.Store(previousConfig)
		slog.SetDefault(previousLogger)
	})
}

func TestForBeforeInitUsesDefaults(t *testing.T) {
	preserveGlobals(t)
	config.Store(nil)

	var output bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))

	For(fallbackTarget{}).Info("fallback message")

	got := output.String()
	for _, want := range []string{
		`"msg":"fallback message"`,
		`"pkg":"github.com/shvydky/logx"`,
		`"type":"fallbackTarget"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestInitNilUsesDefaultConfig(t *testing.T) {
	preserveGlobals(t)

	var output bytes.Buffer
	logger := Init(nil, WithWriter(&output))
	logger.Info("default config message")

	if got := output.String(); !strings.Contains(got, `"msg":"default config message"`) {
		t.Fatalf("unexpected output: %q", got)
	}
}
