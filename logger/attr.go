package logger

import (
	"log/slog"
)

type AnsiString struct {
	Prefix string
	Value  string
}

const badKey = "!BADKEY"

// argsToAttrs is equivalent to slog.argsToAttrSlice().
//
//   - If args[i] is an Attr, it is used as is.
//   - If args[i] is a string and is not the last argument, it is treated as slog.Attr{args[i], args[i+1]}.
//   - Otherwise, the args[i] is treated as slog.Attr{"!BADKEY", args[i]}.
func argsToAttrs(args []any) (attrs []slog.Attr) {
	for i := 0; i < len(args); i++ {
		switch x := args[i].(type) {
		case string:
			if i+1 < len(args) {
				attrs = append(attrs, slog.Any(x, args[i+1]))
				i++
			} else {
				attrs = append(attrs, slog.String(badKey, x))
			}
		case slog.Attr:
			attrs = append(attrs, x)
		default:
			attrs = append(attrs, slog.Any(badKey, x))
		}
	}
	return attrs
}
