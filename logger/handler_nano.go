package logger

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
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
		appendNanoValue(&h2.preformatted, a.Value.Resolve())
	}
	return h2
}

func (h *NanoHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *NanoHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBufferFromPool()
	defer freeBuffer(buf)

	appendDateTime(buf, r.Time)

	*buf = append(*buf, ' ')
	appendNanoLevel(buf, r.Level, h.opts.Colorful)

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

	*buf = append(*buf, ' ')
	*buf = append(*buf, r.Message...)

	if len(h.preformatted) > 0 {
		*buf = append(*buf, h.preformatted...)
	}

	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			appendNanoValue(buf, a.Value.Resolve())
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
	case slog.KindGroup:
		for _, a := range v.Group() {
			appendNanoValue(buf, a.Value.Resolve())
		}
	default:
		*buf = append(*buf, ' ')
		*buf = append(*buf, v.String()...)
	}
}
