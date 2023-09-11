package logger

import "log/slog"

type Options struct {
	AddSource bool
	Colorful  bool
	Level     slog.Level
}
