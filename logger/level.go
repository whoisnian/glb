package logger

import (
	"log/slog"

	"github.com/whoisnian/glb/ansi"
)

const (
	LevelDebug slog.Level = 0  // For debugging in dev environment
	LevelInfo  slog.Level = 6  // For parameters, requests, responses and metrics
	LevelWarn  slog.Level = 12 // For issues that do not affect workflow
	LevelError slog.Level = 18 // For issues that do affect workflow
	LevelFatal slog.Level = 24 // For issues that require immediate termination

	offsetNano = 0
	offsetText = 2
	offsetJson = 4
)

var labelList = []string{
	"[D]", ansi.MagentaFG + "[D]" + ansi.Reset,
	"DEBUG", ansi.MagentaFG + "DEBUG" + ansi.Reset,
	`"DEBUG"`, ansi.MagentaFG + `"DEBUG"` + ansi.Reset,

	"[I]", ansi.GreenFG + "[I]" + ansi.Reset,
	"INFO", ansi.GreenFG + "INFO" + ansi.Reset,
	`"INFO"`, ansi.GreenFG + `"INFO"` + ansi.Reset,

	"[W]", ansi.YellowFG + "[W]" + ansi.Reset,
	"WARN", ansi.YellowFG + "WARN" + ansi.Reset,
	`"WARN"`, ansi.YellowFG + `"WARN"` + ansi.Reset,

	"[E]", ansi.RedFG + "[E]" + ansi.Reset,
	"ERROR", ansi.RedFG + "ERROR" + ansi.Reset,
	`"ERROR"`, ansi.RedFG + `"ERROR"` + ansi.Reset,

	"[F]", ansi.RedFG + "[F]" + ansi.Reset,
	"FATAL", ansi.RedFG + "FATAL" + ansi.Reset,
	`"FATAL"`, ansi.RedFG + `"FATAL"` + ansi.Reset,
}

func appendNanoLevel(buf *[]byte, l slog.Level, colorful bool) {
	if colorful {
		*buf = append(*buf, labelList[l+offsetNano+1]...)
	} else {
		*buf = append(*buf, labelList[l+offsetNano]...)
	}
}

func appendTextLevel(buf *[]byte, l slog.Level, colorful bool) {
	if colorful {
		*buf = append(*buf, labelList[l+offsetText+1]...)
	} else {
		*buf = append(*buf, labelList[l+offsetText]...)
	}
}

func appendJsonLevel(buf *[]byte, l slog.Level, colorful bool) {
	if colorful {
		*buf = append(*buf, labelList[l+offsetJson+1]...)
	} else {
		*buf = append(*buf, labelList[l+offsetJson]...)
	}
}
