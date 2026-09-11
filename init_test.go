package logx

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestForWaitsForConsistentState(t *testing.T) {
	preserveGlobals(t)

	stateMu.Lock()
	var wg sync.WaitGroup
	wg.Add(1)
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer wg.Done()
		close(started)
		For(fallbackTarget{})
		close(done)
	}()
	<-started

	select {
	case <-done:
		stateMu.Unlock()
		wg.Wait()
		t.Fatal("For returned while logger state was being updated")
	case <-time.After(10 * time.Millisecond):
	}

	stateMu.Unlock()
	wg.Wait()
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

func TestForDereferencesNestedPointers(t *testing.T) {
	preserveGlobals(t)

	var output bytes.Buffer
	Init(&Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			"github.com/shvydky/logx.fallbackTarget": slog.LevelDebug,
		},
	}, WithWriter(&output))

	var target *fallbackTarget
	For(&target).Debug("nested pointer message")

	got := output.String()
	for _, want := range []string{
		`"msg":"nested pointer message"`,
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

func TestLevelFor(t *testing.T) {
	lc := newLevelConfig(&Config{
		DefaultLevel: slog.LevelInfo,
		Levels: map[string]slog.Level{
			" example.com/app/worker ":         slog.LevelWarn,
			"example.com/app/worker.Worker":    slog.LevelError,
			"example.com/app/worker.v2":        slog.LevelDebug,
			"example.com/app/worker.v2.Worker": slog.Level(-8),
		},
	})

	tests := []struct {
		name string
		pkg  string
		full string
		want slog.Level
	}{
		{name: "package", pkg: "example.com/app/worker", full: "example.com/app/worker.Other", want: slog.LevelWarn},
		{name: "type takes priority", pkg: "example.com/app/worker", full: "example.com/app/worker.Worker", want: slog.LevelError},
		{name: "package containing dot", pkg: "example.com/app/worker.v2", full: "example.com/app/worker.v2.Other", want: slog.LevelDebug},
		{name: "type in package containing dot", pkg: "example.com/app/worker.v2", full: "example.com/app/worker.v2.Worker", want: slog.Level(-8)},
		{name: "default", pkg: "example.com/app/other", full: "example.com/app/other.Worker", want: slog.LevelInfo},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := lc.levelFor(test.pkg, test.full); got != test.want {
				t.Fatalf("levelFor(%q, %q) = %s, want %s", test.pkg, test.full, got, test.want)
			}
		})
	}
}

func TestLevelConfigResolvesNormalizedKeyCollisionsDeterministically(t *testing.T) {
	lc := newLevelConfig(&Config{
		Levels: map[string]slog.Level{
			" example.com/app ": slog.LevelDebug,
			"example.com/app":   slog.LevelError,
			"  fallback ":       slog.LevelWarn,
			" fallback ":        slog.LevelDebug,
		},
	})

	if got := lc.levels["example.com/app"]; got != slog.LevelError {
		t.Fatalf("exact key level = %s, want %s", got, slog.LevelError)
	}
	if got := lc.levels["fallback"]; got != slog.LevelWarn {
		t.Fatalf("first whitespace variant level = %s, want %s", got, slog.LevelWarn)
	}
}
