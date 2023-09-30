// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bench2

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/whoisnian/glb/logger"
)

// We pass Attrs inline because it affects allocations: building
// up a list outside of the benchmarked code and passing it in with "..."
// reduces measured allocations.

type LOG interface {
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

func BenchmarkAttrs(b *testing.B) {
	ctx := context.Background()
	for _, handler := range []struct {
		name string
		l    LOG
	}{
		{"disabled", slog.New(disabledHandler{})},
		{"async discard", slog.New(newAsyncHandler())},
		{"fastText discard", slog.New(newFastTextHandler(io.Discard))},
		{"Text discard", slog.New(slog.NewTextHandler(io.Discard, nil))},
		{"JSON discard", slog.New(slog.NewJSONHandler(io.Discard, nil))},
		{"nano discard", logger.New(logger.NewNanoHandler(io.Discard, logger.NewOptions(logger.LevelInfo, true, true)))},
		{"text discard", logger.New(logger.NewTextHandler(io.Discard, logger.NewOptions(logger.LevelInfo, true, true)))},
		{"json discard", logger.New(logger.NewJsonHandler(io.Discard, logger.NewOptions(logger.LevelInfo, true, true)))},
	} {
		logg := handler.l
		b.Run(handler.name, func(b *testing.B) {
			for _, call := range []struct {
				name string
				f    func()
			}{
				{
					// The number should match nAttrsInline in slog/record.go.
					// This should exercise the code path where no allocations
					// happen in Record or Attr. If there are allocations, they
					// should only be from Duration.String and Time.String.
					"5 args",
					func() {
						logg.LogAttrs(nil, slog.LevelInfo, testMessage,
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
						)
					},
				},
				{
					"5 args ctx",
					func() {
						logg.LogAttrs(ctx, slog.LevelInfo, testMessage,
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
						)
					},
				},
				{
					"10 args",
					func() {
						logg.LogAttrs(nil, slog.LevelInfo, testMessage,
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
						)
					},
				},
				{
					// Try an extreme value to see if the results are reasonable.
					"40 args",
					func() {
						logg.LogAttrs(nil, slog.LevelInfo, testMessage,
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
							slog.String("string", testString),
							slog.Int("status", testInt),
							slog.Duration("duration", testDuration),
							slog.Time("time", testTime),
							slog.Any("error", testError),
						)
					},
				},
			} {
				b.Run(call.name, func(b *testing.B) {
					b.ReportAllocs()
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							call.f()
						}
					})
				})
			}
		})
	}
}
