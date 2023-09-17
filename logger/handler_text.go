// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
)

type TextHandler struct {
	opts         *Options
	preformatted []byte
	groupPrefix  string

	outMu *sync.Mutex
	out   io.Writer
}

func NewTextHandler(w io.Writer, opts *Options) *TextHandler {
	return &TextHandler{
		opts:  opts,
		outMu: &sync.Mutex{},
		out:   w,
	}
}

func (h *TextHandler) clone() *TextHandler {
	return &TextHandler{
		opts:         h.opts,
		preformatted: slices.Clip(h.preformatted),
		groupPrefix:  h.groupPrefix,
		outMu:        h.outMu,
		out:          h.out,
	}
}

func (h *TextHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.opts.Level
}

func (h *TextHandler) WithAttrs(as []slog.Attr) slog.Handler {
	if len(as) == 0 {
		return h
	}

	h2 := h.clone()
	for _, a := range as {
		appendTextAttr(&h2.preformatted, a, h.groupPrefix)
	}
	return h2
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	if len(h2.groupPrefix) == 0 {
		h2.groupPrefix = name
	} else {
		h2.groupPrefix = h2.groupPrefix + "." + name
	}
	return h2
}

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
	appendFullLevel(buf, r.Level, h.opts.Colorful)
	// source
	if h.opts.AddSource {
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
	if a.Value.Kind() == slog.KindGroup {
		for _, aa := range a.Value.Group() {
			appendTextAttr(buf, aa, prefix+"."+a.Key)
		}
		return
	}

	*buf = append(*buf, ' ')
	appendTextKey(buf, a.Key, prefix)
	appendTextValue(buf, a.Value)
}

func appendTextKey(buf *[]byte, key string, prefix string) {
	if len(prefix) > 0 {
		appendTextString(buf, prefix+"."+key)
	} else {
		appendTextString(buf, key)
	}
	*buf = append(*buf, '=')
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
			return
		} else if vv, ok := va.(error); ok {
			appendTextString(buf, vv.Error())
			return
		} else if vv, ok := va.([]byte); ok {
			appendTextString(buf, string(vv))
			return
		}
		appendTextString(buf, fmt.Sprint(va))
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

	for _, r := range str {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			*buf = strconv.AppendQuote(*buf, str)
			return
		}
	}
	*buf = append(*buf, str...)
}
