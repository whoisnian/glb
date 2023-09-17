package logger

import (
	"log/slog"

	"github.com/whoisnian/glb/ansi"
)

const (
	LevelDebug slog.Level = 0  // For debugging in dev environment
	LevelInfo  slog.Level = 4  // For parameters, requests, responses and metrics
	LevelWarn  slog.Level = 8  // For issues that do not affect workflow
	LevelError slog.Level = 12 // For issues that do affect workflow
	LevelFatal slog.Level = 16 // For issues that require immediate termination
)

func ValidLevel(l slog.Level) bool {
	return l&3 == 0 && l >= 0 && l <= 16
}

var labelList = []string{
	"[D]", ansi.MagentaFG + "[D]" + ansi.Reset,
	"DEBUG", ansi.MagentaFG + "DEBUG" + ansi.Reset,

	"[I]", ansi.GreenFG + "[I]" + ansi.Reset,
	"INFO", ansi.GreenFG + "INFO" + ansi.Reset,

	"[W]", ansi.YellowFG + "[W]" + ansi.Reset,
	"WARN", ansi.YellowFG + "WARN" + ansi.Reset,

	"[E]", ansi.RedFG + "[E]" + ansi.Reset,
	"ERROR", ansi.RedFG + "ERROR" + ansi.Reset,

	"[F]", ansi.RedFG + "[F]" + ansi.Reset,
	"FATAL", ansi.RedFG + "FATAL" + ansi.Reset,
}

func appendShortLevel(buf *[]byte, l slog.Level, colorful bool) {
	if colorful {
		*buf = append(*buf, labelList[l+1]...)
	} else {
		*buf = append(*buf, labelList[l]...)
	}
}

func appendFullLevel(buf *[]byte, l slog.Level, colorful bool) {
	if colorful {
		*buf = append(*buf, labelList[l+3]...)
	} else {
		*buf = append(*buf, labelList[l+2]...)
	}
}
