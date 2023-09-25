package logger

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// TextHandler formats slog.Record as a sequence of key=value pairs separated by spaces and followed by a newline.
type TextHandler struct {
	*Options
	outMu *sync.Mutex
	out   io.Writer

	preformatted []byte
	groupPrefix  string
}

// NewTextHandler creates a new TextHandler with the given io.Writer and Options.
// The Options should not be changed after first use.
func NewTextHandler(w io.Writer, opts *Options) *TextHandler {
	return &TextHandler{
		Options: opts,
		outMu:   &sync.Mutex{},
		out:     w,
	}
}

func (h *TextHandler) clone() *TextHandler {
	return &TextHandler{
		Options:      h.Options,
		outMu:        h.outMu,
		out:          h.out,
		preformatted: slices.Clip(h.preformatted),
		groupPrefix:  h.groupPrefix,
	}
}

// WithAttrs returns a new TextHandler whose attributes consists of h's attributes followed by attrs.
// If attrs is empty, WithAttrs returns the origin TextHandler.
func (h *TextHandler) WithAttrs(attrs []slog.Attr) Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()
	for _, a := range attrs {
		appendTextAttr(&h2.preformatted, a, h.groupPrefix)
	}
	return h2
}

// WithGroup returns a new TextHandler that starts a group with the given name.
// If name is empty, WithGroup returns the origin TextHandler.
func (h *TextHandler) WithGroup(name string) Handler {
	h2 := h.clone()
	if len(h2.groupPrefix) == 0 {
		h2.groupPrefix = name
	} else {
		h2.groupPrefix = h2.groupPrefix + "." + name
	}
	return h2
}

// Handle formats slog.Record as a single line of space-separated key=value items.
//
// The time is output in [time.RFC3339] format.
//
// Keys and values are quoted with [strconv.Quote] if they contain Unicode space
// characters, non-printing characters, '"' or '='.
func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBuffer()
	defer freeBuffer(buf)

	// time
	*buf = append(*buf, slog.TimeKey...)
	*buf = append(*buf, '=')
	*buf = r.Time.AppendFormat(*buf, time.RFC3339)
	// level
	*buf = append(*buf, ' ')
	*buf = append(*buf, slog.LevelKey...)
	*buf = append(*buf, '=')
	appendFullLevel(buf, r.Level, h.Options.colorful)
	// source
	if h.Options.addSource {
		*buf = append(*buf, ' ')
		*buf = append(*buf, slog.SourceKey...)
		*buf = append(*buf, '=')
		appendTextSource(buf, r.PC)
	}
	// msg
	*buf = append(*buf, ' ')
	*buf = append(*buf, slog.MessageKey...)
	*buf = append(*buf, '=')
	appendTextString(buf, r.Message)

	if len(h.preformatted) > 0 {
		*buf = append(*buf, h.preformatted...)
	}

	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			appendTextAttr(buf, a, h.groupPrefix)
			return true
		})
	}
	*buf = append(*buf, '\n')

	h.outMu.Lock()
	defer h.outMu.Unlock()
	_, err := h.out.Write(*buf)
	return err
}

func appendTextAttr(buf *[]byte, a slog.Attr, prefix string) {
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		for _, aa := range a.Value.Group() {
			if len(prefix) > 0 {
				appendTextAttr(buf, aa, prefix+"."+a.Key)
			} else {
				appendTextAttr(buf, aa, a.Key)
			}
		}
		return
	}

	*buf = append(*buf, ' ')
	if len(prefix) > 0 {
		appendTextString(buf, prefix+"."+a.Key)
	} else {
		appendTextString(buf, a.Key)
	}
	*buf = append(*buf, '=')
	appendTextValue(buf, a.Value)
}

func appendTextValue(buf *[]byte, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		appendTextString(buf, v.String())
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
		va := v.Any()
		if vv, ok := va.(encoding.TextMarshaler); ok {
			if data, err := vv.MarshalText(); err != nil {
				appendTextString(buf, err.Error())
			} else {
				appendTextString(buf, string(data))
			}
		} else if vv, ok := va.(error); ok {
			appendTextString(buf, vv.Error())
		} else if vv, ok := va.([]byte); ok {
			appendTextString(buf, string(vv))
		} else {
			appendTextString(buf, fmt.Sprint(va))
		}
	}
}

func appendTextSource(buf *[]byte, pc uintptr) {
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
	appendTextString(buf, f.File[idx+1:]+":"+strconv.FormatInt(int64(f.Line), 10))
}

func appendTextString(buf *[]byte, str string) {
	if len(str) == 0 {
		*buf = append(*buf, '"', '"')
		return
	}

	for i := 0; i < len(str); {
		b := str[i]
		if b < utf8.RuneSelf {
			if b != '\\' && (b == ' ' || b == '=' || !safeSet[b]) {
				*buf = strconv.AppendQuote(*buf, str)
				return
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(str[i:])
		if r == utf8.RuneError || unicode.IsSpace(r) || !unicode.IsPrint(r) {
			*buf = strconv.AppendQuote(*buf, str)
			return
		}
		i += size
	}
	*buf = append(*buf, str...)
}
