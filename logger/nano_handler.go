package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"
)

// NanoHandler formats slog.Record as a sequence of value strings without attribute keys to minimize log length.
type NanoHandler struct {
	*Options
	outMu *sync.Mutex
	out   io.Writer

	preformatted []byte
}

// NewNanoHandler creates a new NanoHandler with the given io.Writer and Options.
// The Options should not be changed after first use.
func NewNanoHandler(w io.Writer, opts *Options) *NanoHandler {
	return &NanoHandler{
		Options: opts,
		outMu:   &sync.Mutex{},
		out:     w,
	}
}

func (h *NanoHandler) clone() *NanoHandler {
	return &NanoHandler{
		Options:      h.Options,
		outMu:        h.outMu,
		out:          h.out,
		preformatted: slices.Clip(h.preformatted),
	}
}

// WithAttrs returns a new NanoHandler whose attributes consists of h's attributes followed by attrs.
// If attrs is empty, WithAttrs returns the origin NanoHandler.
func (h *NanoHandler) WithAttrs(attrs []slog.Attr) Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()
	for _, a := range attrs {
		appendNanoValue(&h2.preformatted, a.Value)
	}
	return h2
}

// WithGroup returns the origin NanoHandler because NanoHandler always ignores attribute keys.
func (h *NanoHandler) WithGroup(name string) Handler {
	return h
}

// Handle formats slog.Record as a sequence of value strings without attribute keys.
//
// The time is output in [time.DateTime] format.
//
// If the Record's message is empty, the message is omitted.
func (h *NanoHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBuffer()
	defer freeBuffer(buf)

	// time
	appendDateTime(buf, r.Time)
	// level
	*buf = append(*buf, ' ')
	appendShortLevel(buf, r.Level, h.Options.colorful)
	// source
	if h.Options.addSource {
		*buf = append(*buf, ' ')
		appendNanoSource(buf, r.PC)
	}
	// msg
	if len(r.Message) > 0 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, r.Message...)
	}

	if len(h.preformatted) > 0 {
		*buf = append(*buf, h.preformatted...)
	}

	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			appendNanoValue(buf, a.Value)
			return true
		})
	}
	*buf = append(*buf, '\n')

	h.outMu.Lock()
	defer h.outMu.Unlock()
	_, err := h.out.Write(*buf)
	return err
}

func appendNanoValue(buf *[]byte, v slog.Value) {
	v = v.Resolve()
	if v.Kind() == slog.KindGroup {
		for _, a := range v.Group() {
			appendNanoValue(buf, a.Value)
		}
		return
	}

	*buf = append(*buf, ' ')
	switch v.Kind() {
	case slog.KindString:
		*buf = append(*buf, v.String()...)
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		*buf = append(*buf, v.Duration().String()...)
	case slog.KindTime:
		*buf = v.Time().AppendFormat(*buf, time.RFC3339)
	case slog.KindAny, slog.KindLogValuer:
		*buf = fmt.Append(*buf, v.Any())
	}
}

func appendNanoSource(buf *[]byte, pc uintptr) {
	f, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	idx, first := 0, false
	for idx = len(f.File) - 1; idx > 0; idx-- {
		if f.File[idx] == '/' {
			if first {
				break
			}
			first = true
		}
	}
	*buf = append(*buf, f.File[idx+1:]...)
	*buf = append(*buf, ':')
	*buf = strconv.AppendInt(*buf, int64(f.Line), 10)
}

func appendDateTime(buf *[]byte, t time.Time) {
	year, month, day := t.Date()
	appendIntWidth4(buf, year)
	*buf = append(*buf, '-')
	appendIntWidth2(buf, int(month))
	*buf = append(*buf, '-')
	appendIntWidth2(buf, day)
	*buf = append(*buf, ' ')

	hour, min, sec := t.Clock()
	appendIntWidth2(buf, hour)
	*buf = append(*buf, ':')
	appendIntWidth2(buf, min)
	*buf = append(*buf, ':')
	appendIntWidth2(buf, sec)
}
