package logger

import (
	"context"
	"log/slog"
)

// Options is the common options for all handlers.
type Options struct {
	level     slog.Level
	colorful  bool
	addSource bool
}

// NewOptions creates a new Options with the given level, colorful and addSource.
func NewOptions(level slog.Level, colorful bool, addSource bool) *Options {
	return &Options{level, colorful, addSource}
}

// Enabled reports whether the given level is enabled.
func (opts *Options) Enabled(l slog.Level) bool {
	return l >= opts.level
}

// IsDebug reports whether the handler is LevelDebug.
func (opts *Options) IsDebug() bool {
	return LevelDebug == opts.level
}

// IsColorful reports whether the handler enables colorful level labels.
func (opts *Options) IsColorful() bool {
	return opts.colorful
}

// IsAddSource reports whether the handler adds source info.
func (opts *Options) IsAddSource() bool {
	return opts.addSource
}

// Handler handles log records produced by a [Logger].
//
// Users should use the methods of Logger instead of invoking Handler methods directly.
type Handler interface {
	Enabled(slog.Level) bool
	IsDebug() bool
	IsColorful() bool
	IsAddSource() bool

	WithAttrs(attrs []slog.Attr) Handler
	WithGroup(name string) Handler
	Handle(context.Context, slog.Record) error
}
