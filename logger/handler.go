package logger

import (
	"log/slog"
	"reflect"
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
	result, _ = extractHandlerOptions(reflect.ValueOf(h))
	return result
}

func extractHandlerOptions(v reflect.Value) (result bool, ok bool) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		vt := v.Type()
		if vt.Name() == "HandlerOptions" && vt.PkgPath() == "log/slog" {
			if vv := v.FieldByName("AddSource"); vv.Kind() == reflect.Bool {
				return vv.Bool(), true
			}
		}
		for i := 0; i < v.NumField(); i++ {
			if result, ok = extractHandlerOptions(v.Field(i)); ok {
				return result, true
			}
		}
	}
	return true, false
}
