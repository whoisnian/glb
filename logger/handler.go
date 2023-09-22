package logger

import (
	"context"
	"log/slog"
)

type Options struct {
	level     slog.Level
	colorful  bool
	addSource bool
}

func NewOptions(level slog.Level, colorful bool, addSource bool) *Options {
	return &Options{level, colorful, addSource}
}

func (opts *Options) Enabled(l slog.Level) bool {
	return l >= opts.level
}

func (opts *Options) IsDebug() bool {
	return LevelDebug == opts.level
}

func (opts *Options) IsColorful() bool {
	return opts.colorful
}

func (opts *Options) IsAddSource() bool {
	return opts.addSource
}

type Handler interface {
	Enabled(slog.Level) bool
	IsDebug() bool
	IsColorful() bool
	IsAddSource() bool

	WithAttrs(attrs []slog.Attr) Handler
	WithGroup(name string) Handler
	Handle(context.Context, slog.Record) error
}
