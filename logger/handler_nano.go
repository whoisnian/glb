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

type NanoHandler struct {
	opts         *Options
	preformatted []byte

	outMu *sync.Mutex
	out   io.Writer
}

func NewNanoHandler(w io.Writer, opts *Options) *NanoHandler {
	return &NanoHandler{
		opts:  opts,
		outMu: &sync.Mutex{},
		out:   w,
	}
}

func (h *NanoHandler) clone() *NanoHandler {
	return &NanoHandler{
		opts:         h.opts,
		preformatted: slices.Clip(h.preformatted),
		outMu:        h.outMu,
		out:          h.out,
	}
}

func (h *NanoHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.opts.Level
}

func (h *NanoHandler) WithAttrs(as []slog.Attr) slog.Handler {
	h2 := h.clone()
	for _, a := range as {
		appendNanoValue(&h2.preformatted, a.Value)
	}
	return h2
}

func (h *NanoHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *NanoHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBuffer()
	defer freeBuffer(buf)

	// time
	appendDateTime(buf, r.Time)
	// level
	*buf = append(*buf, ' ')
	appendNanoLevel(buf, r.Level, h.opts.Colorful)
	// source
	if h.opts.AddSource {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		idx, first := 0, false
		for idx = len(f.File) - 1; idx > 0; idx-- {
			if f.File[idx] == '/' {
				if first {
					break
				}
				first = true
			}
		}
		*buf = append(*buf, ' ')
		*buf = append(*buf, f.File[idx+1:]...)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(f.Line), 10)
	}
	// msg
	*buf = append(*buf, ' ')
	*buf = append(*buf, r.Message...)

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
	switch v.Kind() {
	case slog.KindString:
		*buf = append(*buf, ' ')
		*buf = append(*buf, v.String()...)
	case slog.KindInt64:
		*buf = append(*buf, ' ')
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = append(*buf, ' ')
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = append(*buf, ' ')
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = append(*buf, ' ')
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		*buf = append(*buf, ' ')
		*buf = append(*buf, v.Duration().String()...)
	case slog.KindTime:
		*buf = append(*buf, ' ')
		*buf = v.Time().AppendFormat(*buf, time.RFC3339Nano)
	case slog.KindGroup:
		for _, a := range v.Group() {
			appendNanoValue(buf, a.Value)
		}
	case slog.KindAny, slog.KindLogValuer:
		*buf = append(*buf, ' ')
		*buf = fmt.Append(*buf, v.Any())
	}
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
