package logger

import (
	"log/slog"
	"testing"
)

func TestOptionsEnabled(t *testing.T) {
	var tests = []struct {
		opts  Options
		input slog.Level
		want  bool
	}{
		{Options{level: LevelDebug}, LevelDebug, true},
		{Options{level: LevelDebug}, LevelInfo, true},
		{Options{level: LevelDebug}, LevelWarn, true},
		{Options{level: LevelInfo}, LevelDebug, false},
		{Options{level: LevelInfo}, LevelInfo, true},
		{Options{level: LevelInfo}, LevelError, true},
		{Options{level: LevelError}, LevelDebug, false},
		{Options{level: LevelError}, LevelWarn, false},
		{Options{level: LevelError}, LevelError, true},
		{Options{level: LevelError}, LevelFatal, true},
	}
	for _, test := range tests {
		got := test.opts.Enabled(test.input)
		if got != test.want {
			t.Errorf("%v.Enabled(%d) = %v, want %v", test.opts, test.input, got, test.want)
		}
	}
}

func TestOptionsCheck(t *testing.T) {
	var tests = []struct {
		opts  *Options
		wantD bool
		wantC bool
		wantA bool
	}{
		{NewOptions(LevelDebug, false, false), true, false, false},
		{NewOptions(LevelDebug, true, false), true, true, false},
		{NewOptions(LevelInfo, true, true), false, true, true},
		{NewOptions(LevelWarn, false, false), false, false, false},
		{NewOptions(LevelError, false, true), false, false, true},
		{NewOptions(LevelFatal, true, true), false, true, true},
	}
	for _, test := range tests {
		if test.opts.IsDebug() != test.wantD {
			t.Errorf("%v.IsDebug() = %v, want %v", test.opts, test.opts.IsDebug(), test.wantD)
		}
		if test.opts.IsColorful() != test.wantC {
			t.Errorf("%v.IsColorful() = %v, want %v", test.opts, test.opts.IsColorful(), test.wantC)
		}
		if test.opts.IsAddSource() != test.wantA {
			t.Errorf("%v.IsAddSource() = %v, want %v", test.opts, test.opts.IsAddSource(), test.wantA)
		}
	}
}
