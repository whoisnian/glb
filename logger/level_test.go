package logger

import (
	"log/slog"
	"regexp"
	"slices"
	"testing"
)

var allLevels = []slog.Level{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}

func TestValidLevel(t *testing.T) {
	for i := -256; i <= 256; i++ {
		l := slog.Level(i)
		if ValidLevel(l) && !slices.Contains(allLevels, l) {
			t.Errorf("ValidLevel(%d) = true, want false", i)
		}
	}
}

var (
	reShort         = regexp.MustCompile(`^\[(D|I|W|E|F)\]$`)
	reShortColorful = regexp.MustCompile(`^\x1b\[3[1235]m\[(D|I|W|E|F)\]\x1b\[0m$`)
	reFull          = regexp.MustCompile(`^(DEBUG|INFO|WARN|ERROR|FATAL)$`)
	reFullColorful  = regexp.MustCompile(`^\x1b\[3[1235]m(DEBUG|INFO|WARN|ERROR|FATAL)\x1b\[0m$`)
)

func TestAppendShortLevel(t *testing.T) {
	buf := make([]byte, 32)
	for _, l := range allLevels {
		buf = buf[:0]
		if appendShortLevel(&buf, l, false); !reShort.Match(buf) {
			t.Fatalf("short level should match %q is %q", reShort, buf)
		}
	}
	for _, l := range allLevels {
		buf = buf[:0]
		if appendShortLevel(&buf, l, true); !reShortColorful.Match(buf) {
			t.Fatalf("short colorful level should match %q is %q", reShortColorful, buf)
		}
	}
}

func TestAppendFullLevel(t *testing.T) {
	buf := make([]byte, 32)
	for _, l := range allLevels {
		buf = buf[:0]
		if appendFullLevel(&buf, l, false); !reFull.Match(buf) {
			t.Fatalf("full level should match %q is %q", reFull, buf)
		}
	}
	for _, l := range allLevels {
		buf = buf[:0]
		if appendFullLevel(&buf, l, true); !reFullColorful.Match(buf) {
			t.Fatalf("full colorful level should match %q is %q", reFullColorful, buf)
		}
	}
}
