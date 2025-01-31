package logger

import (
	"log/slog"
	"reflect"
	"strings"
)

// Options is the common options for all handlers.
type Options struct {
	Level     slog.Level
	Colorful  bool
	AddSource bool
}

func tryIsAddSource(h slog.Handler) (result bool) {
	if hh, ok := h.(interface{ IsAddSource() bool }); ok {
		return hh.IsAddSource()
	}
	result, _ = extractHandlerOptions(reflect.ValueOf(h), 16)
	return result
}

func extractHandlerOptions(v reflect.Value, depth int) (result bool, ok bool) {
	for d := 16; d > 0 && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface); d-- {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		matchFunc := func(s string) bool { return strings.ToLower(s) == "addsource" }
		if vv := v.FieldByNameFunc(matchFunc); vv.IsValid() && vv.Kind() == reflect.Bool {
			return vv.Bool(), true
		}

		for i := 0; depth > 0 && i < v.NumField(); i++ {
			if result, ok = extractHandlerOptions(v.Field(i), depth-1); ok {
				return result, true
			}
		}
	}
	return true, false
}
