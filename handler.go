package logx

import (
	"context"
	"log/slog"
)

type levelHandler struct {
	next  slog.Handler
	level slog.Level
}

func newLevelHandler(next slog.Handler, level slog.Level) slog.Handler {
	return &levelHandler{next: next, level: level}
}

func (h *levelHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	if lvl < h.level {
		return false
	}
	return h.next.Enabled(ctx, lvl)
}

func (h *levelHandler) Handle(ctx context.Context, rec slog.Record) error {
	if rec.Level < h.level {
		return nil
	}

	return h.next.Handle(ctx, rec)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{next: h.next.WithAttrs(attrs), level: h.level}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{next: h.next.WithGroup(name), level: h.level}
}
