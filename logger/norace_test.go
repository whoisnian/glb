//go:build !race

package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestNanoHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := range 10 {
		r.AddAttrs(slog.Int("x = y", i))
	}
	var h slog.Handler = NewNanoHandler(io.Discard, Options{LevelInfo, false, false})
	got := int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("origin.Handle() got %d allocs, want 0", got)
	}

	h = h.WithGroup("s")
	r.AddAttrs(slog.Group("g", slog.Int("a", 1)))
	got = int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("new.Handle() got %d allocs, want 0", got)
	}
}

func TestTextHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := range 10 {
		r.AddAttrs(slog.Int("x = y", i))
	}
	var h slog.Handler = NewTextHandler(io.Discard, Options{LevelInfo, false, false})
	got := int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("origin.Handle() got %d allocs, want 0", got)
	}

	h = h.WithGroup("s")
	r.AddAttrs(slog.Group("g", slog.Int64("a", 1)))
	got = int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("new.Handle() got %d allocs, want 0", got)
	}
}

func TestJsonHandlerAlloc(t *testing.T) {
	r := slog.NewRecord(time.Now(), LevelInfo, "msg", 0)
	for i := range 10 {
		r.AddAttrs(slog.Int("x", i))
	}
	var h slog.Handler = NewJsonHandler(io.Discard, Options{LevelInfo, false, false})
	got := int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("origin.Handle() got %d allocs, want 0", got)
	}

	h = h.WithGroup("s")
	r.AddAttrs(slog.Group("g", slog.Int64("a", 1)))
	got = int(testing.AllocsPerRun(5, func() { h.Handle(context.Background(), r) }))
	if got != 0 {
		t.Errorf("new.Handle() got %d allocs, want 0", got)
	}
}
