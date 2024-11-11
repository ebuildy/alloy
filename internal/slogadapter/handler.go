package slogadapter

import (
	"context"
	"log/slog"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var _ slog.Handler = (*GoKitHandler)(nil)

type GoKitHandler struct {
	Logger log.Logger
	Group  string
}

func (h GoKitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h GoKitHandler) Handle(ctx context.Context, record slog.Record) error {
	var logger log.Logger
	switch record.Level {
	case slog.LevelInfo:
		logger = level.Info(h.Logger)
	case slog.LevelWarn:
		logger = level.Warn(h.Logger)
	case slog.LevelError:
		logger = level.Error(h.Logger)
	default:
		logger = level.Debug(h.Logger)
	}

	if h.Group == "" {
		pairs := make([]any, 0, record.NumAttrs()+2)
		pairs = append(pairs, "msg", record.Message)
		record.Attrs(func(attr slog.Attr) bool {
			pairs = append(pairs, attr.Key, attr.Value)
			return true
		})
		return logger.Log(pairs...)
	}

	pairs := make([]any, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		pairs = append(pairs, attr.Key, attr.Value)
		return true
	})
	g := slog.Group(h.Group, pairs...)
	pairs = []any{"msg", record.Message, g.Key, g.Value}
	return logger.Log(pairs...)
}

func (h GoKitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	pairs := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		pairs = append(pairs, attr.Key, attr.Value)
	}

	if h.Group == "" {
		return GoKitHandler{Logger: log.With(h.Logger, pairs...)}
	}

	g := slog.Group(h.Group, pairs...)
	return GoKitHandler{Logger: log.With(h.Logger, g.Key, g.Value)}
}

func (h GoKitHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h.Group = name
	return h
}
