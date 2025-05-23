package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/whoisnian/glb/ansi"
)

// JsonHandler formats slog.Record as line-delimited JSON objects.
type JsonHandler struct {
	level     slog.Level
	colorful  bool
	addSource bool

	outMu *sync.Mutex
	out   io.Writer

	preformatted []byte
	nOpenGroups  int
	addSep       bool
}

// NewJsonHandler creates a new JsonHandler with the given io.Writer and Options.
// The Options should not be changed after first use.
func NewJsonHandler(w io.Writer, opts Options) *JsonHandler {
	return &JsonHandler{
		level:     opts.Level,
		colorful:  opts.Colorful,
		addSource: opts.AddSource,
		outMu:     &sync.Mutex{},
		out:       w,
		addSep:    true,
	}
}

func (h *JsonHandler) clone() *JsonHandler {
	return &JsonHandler{
		level:        h.level,
		colorful:     h.colorful,
		addSource:    h.addSource,
		outMu:        h.outMu,
		out:          h.out,
		preformatted: slices.Clip(h.preformatted),
		nOpenGroups:  h.nOpenGroups,
		addSep:       h.addSep,
	}
}

// Enabled reports whether the given level is enabled.
func (h *JsonHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

// IsAddSource reports whether the handler adds source info.
func (h *JsonHandler) IsAddSource() bool {
	return h.addSource
}

// WithAttrs returns a new JsonHandler whose attributes consists of h's attributes followed by attrs.
// If attrs is empty, WithAttrs returns the origin JsonHandler.
func (h *JsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()
	for _, a := range attrs {
		appendJsonAttr(&h2.preformatted, a, h2.addSep, h2.colorful)
		h2.addSep = true
	}
	return h2
}

// WithGroup returns a new JsonHandler that starts a group with the given name.
// If name is empty, WithGroup returns the origin JsonHandler.
func (h *JsonHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	if h2.addSep {
		h2.preformatted = append(h2.preformatted, ',', '"')
	} else {
		h2.preformatted = append(h2.preformatted, '"')
	}
	appendJsonString(&h2.preformatted, name)
	h2.preformatted = append(h2.preformatted, '"', ':', '{')
	h2.nOpenGroups += 1
	h2.addSep = false
	return h2
}

// Handle formats slog.Record as a JSON object on a single line.
//
// The time is output in [time.RFC3339Nano] format.
//
// An encoding failure does not cause Handle to return an error.
// Instead, the error message is formatted as a string.
func (h *JsonHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBuffer()
	defer freeBuffer(buf)

	*buf = append(*buf, '{')
	// time
	*buf = append(*buf, '"')
	*buf = append(*buf, slog.TimeKey...)
	*buf = append(*buf, '"', ':', '"')
	*buf = r.Time.AppendFormat(*buf, time.RFC3339Nano)
	// level
	*buf = append(*buf, '"', ',', '"')
	*buf = append(*buf, slog.LevelKey...)
	*buf = append(*buf, '"', ':', '"')
	appendFullLevel(buf, r.Level, h.colorful)
	*buf = append(*buf, '"')
	// source
	if h.addSource {
		*buf = append(*buf, ',', '"')
		*buf = append(*buf, slog.SourceKey...)
		*buf = append(*buf, '"', ':', '{')
		appendJsonSource(buf, r.PC)
		*buf = append(*buf, '}')
	}
	// msg
	*buf = append(*buf, ',', '"')
	*buf = append(*buf, slog.MessageKey...)
	*buf = append(*buf, '"', ':', '"')
	appendJsonString(buf, r.Message)
	*buf = append(*buf, '"')

	if len(h.preformatted) > 0 {
		*buf = append(*buf, h.preformatted...)
	}

	if r.NumAttrs() > 0 {
		addSep := h.addSep
		r.Attrs(func(a slog.Attr) bool {
			appendJsonAttr(buf, a, addSep, h.colorful)
			addSep = true
			return true
		})
	}
	for range h.nOpenGroups {
		*buf = append(*buf, '}')
	}
	*buf = append(*buf, '}', '\n')

	h.outMu.Lock()
	defer h.outMu.Unlock()
	_, err := h.out.Write(*buf)
	return err
}

func appendJsonAttr(buf *[]byte, a slog.Attr, addSep bool, colorful bool) {
	if addSep {
		*buf = append(*buf, ',')
		addSep = false
	}

	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		if len(a.Key) > 0 {
			*buf = append(*buf, '"')
			appendJsonString(buf, a.Key)
			*buf = append(*buf, '"', ':', '{')
		}
		for _, aa := range a.Value.Group() {
			appendJsonAttr(buf, aa, addSep, colorful)
			addSep = true
		}
		if len(a.Key) > 0 {
			*buf = append(*buf, '}')
		}
		return
	}

	*buf = append(*buf, '"')
	appendJsonString(buf, a.Key)
	*buf = append(*buf, '"', ':')
	appendJsonValue(buf, a.Value, colorful)
}

func appendJsonValue(buf *[]byte, v slog.Value, colorful bool) {
	switch v.Kind() {
	case slog.KindString:
		*buf = append(*buf, '"')
		appendJsonString(buf, v.String())
		*buf = append(*buf, '"')
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		appendJsonMarshal(buf, v.Float64())
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		*buf = strconv.AppendInt(*buf, int64(v.Duration()), 10)
	case slog.KindTime:
		*buf = append(*buf, '"')
		*buf = v.Time().AppendFormat(*buf, time.RFC3339Nano)
		*buf = append(*buf, '"')
	case slog.KindAny, slog.KindLogValuer:
		va := v.Any()
		if _, ok := va.(json.Marshaler); ok {
			appendJsonMarshal(buf, va)
		} else if vv, ok := va.(error); ok {
			*buf = append(*buf, '"')
			if colorful {
				*buf = append(*buf, ansi.RedFG...)
				appendJsonString(buf, vv.Error())
				*buf = append(*buf, ansi.Reset...)
			} else {
				appendJsonString(buf, vv.Error())
			}
			*buf = append(*buf, '"')
		} else if vv, ok := va.(AnsiString); ok {
			*buf = append(*buf, '"')
			if colorful && vv.Prefix != "" {
				*buf = append(*buf, vv.Prefix...)
				appendJsonString(buf, vv.Value)
				*buf = append(*buf, ansi.Reset...)
			} else {
				appendJsonString(buf, vv.Value)
			}
			*buf = append(*buf, '"')
		} else {
			appendJsonMarshal(buf, va)
		}
	}
}

func appendJsonMarshal(buf *[]byte, v any) {
	bb := bytes.Buffer{}
	enc := json.NewEncoder(&bb)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		*buf = append(*buf, '"')
		appendJsonString(buf, "!ERROR:"+err.Error())
		*buf = append(*buf, '"')
		return
	}
	bs := bb.Bytes()
	*buf = append(*buf, bs[:len(bs)-1]...)
}

func appendJsonSource(buf *[]byte, pc uintptr) {
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
	*buf = append(*buf, '"', 'f', 'i', 'l', 'e', '"', ':', '"')
	appendJsonString(buf, f.File[idx+1:])
	*buf = append(*buf, '"', ',', '"', 'l', 'i', 'n', 'e', '"', ':')
	*buf = strconv.AppendInt(*buf, int64(f.Line), 10)
}

func appendJsonString(buf *[]byte, str string) {
	start := 0
	for i := 0; i < len(str); {
		if b := str[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			*buf = append(*buf, str[start:i]...)
			switch b {
			case '\\', '"':
				*buf = append(*buf, '\\', b)
			case '\n':
				*buf = append(*buf, '\\', 'n')
			case '\r':
				*buf = append(*buf, '\\', 'r')
			case '\t':
				*buf = append(*buf, '\\', 't')
			default:
				*buf = append(*buf, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(str[i:])
		if c == utf8.RuneError && size == 1 {
			*buf = append(*buf, str[start:i]...)
			*buf = append(*buf, '\\', 'u', 'f', 'f', 'f', 'd')
			i += size
			start = i
			continue
		}
		if c == '\u2028' || c == '\u2029' {
			*buf = append(*buf, str[start:i]...)
			*buf = append(*buf, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	*buf = append(*buf, str[start:]...)
}

const hex = "0123456789abcdef"

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
