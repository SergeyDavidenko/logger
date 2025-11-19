package benchmarks

import (
	"context"
	"io"
	"log/slog"
	"testing"

	apexlog "github.com/apex/log"
	apexloghandler "github.com/apex/log/handlers/discard"
	kitlog "github.com/go-kit/log"
	"github.com/inconshreveable/log15"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/SergeyDavidenko/logger"
)

// Benchmark for this logger with 10 fields
func BenchmarkLogger_With10Fields(b *testing.B) {
	config := logger.Config{
		LogstashEnabled: false,
		AppName:         "bench-app",
		MinLevel:        logger.INFO,
		AsyncEnabled:    false, // Disable async for fair comparison
	}

	l := logger.New(config)
	defer l.Close()
	l.SetOutput(io.Discard)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message: field1=%s field2=%d field3=%s field4=%d field5=%s field6=%d field7=%s field8=%d field9=%s field10=%d",
			"value1", 1, "value2", 2, "value3", 3, "value4", 4, "value5", 5)
	}
}

// Benchmark for zap
func BenchmarkZap_With10Fields(b *testing.B) {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	l := zap.New(core)
	defer l.Sync()
	l = l.With(
		zap.String("field1", "value1"),
		zap.Int("field2", 1),
		zap.String("field3", "value2"),
		zap.Int("field4", 2),
		zap.String("field5", "value3"),
		zap.Int("field6", 3),
		zap.String("field7", "value4"),
		zap.Int("field8", 4),
		zap.String("field9", "value5"),
		zap.Int("field10", 5),
	)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for zap (sugared)
func BenchmarkZapSugared_With10Fields(b *testing.B) {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	l := zap.New(core).Sugar()
	defer l.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Infow("test message",
			"field1", "value1",
			"field2", 1,
			"field3", "value2",
			"field4", 2,
			"field5", "value3",
			"field6", 3,
			"field7", "value4",
			"field8", 4,
			"field9", "value5",
			"field10", 5,
		)
	}
}

// Benchmark for zerolog
func BenchmarkZerolog_With10Fields(b *testing.B) {
	l := zerolog.New(io.Discard).With().
		Str("field1", "value1").
		Int("field2", 1).
		Str("field3", "value2").
		Int("field4", 2).
		Str("field5", "value3").
		Int("field6", 3).
		Str("field7", "value4").
		Int("field8", 4).
		Str("field9", "value5").
		Int("field10", 5).
		Logger()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("test message")
	}
}

// Benchmark for go-kit
func BenchmarkGoKit_With10Fields(b *testing.B) {
	l := kitlog.NewJSONLogger(io.Discard)
	l = kitlog.With(l,
		"field1", "value1",
		"field2", 1,
		"field3", "value2",
		"field4", 2,
		"field5", "value3",
		"field6", 3,
		"field7", "value4",
		"field8", 4,
		"field9", "value5",
		"field10", 5,
	)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Log("msg", "test message")
	}
}

// Benchmark for slog (LogAttrs)
func BenchmarkSlogLogAttrs_With10Fields(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil))

	b.ResetTimer()
	b.ReportAllocs()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		l.LogAttrs(ctx, slog.LevelInfo, "test message",
			slog.String("field1", "value1"),
			slog.Int("field2", 1),
			slog.String("field3", "value2"),
			slog.Int("field4", 2),
			slog.String("field5", "value3"),
			slog.Int("field6", 3),
			slog.String("field7", "value4"),
			slog.Int("field8", 4),
			slog.String("field9", "value5"),
			slog.Int("field10", 5),
		)
	}
}

// Benchmark for slog
func BenchmarkSlog_With10Fields(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message",
			"field1", "value1",
			"field2", 1,
			"field3", "value2",
			"field4", 2,
			"field5", "value3",
			"field6", 3,
			"field7", "value4",
			"field8", 4,
			"field9", "value5",
			"field10", 5,
		)
	}
}

// Benchmark for apex/log
func BenchmarkApexLog_With10Fields(b *testing.B) {
	apexlog.SetHandler(apexloghandler.Default)
	l := apexlog.WithFields(apexlog.Fields{
		"field1":  "value1",
		"field2":  1,
		"field3":  "value2",
		"field4":  2,
		"field5":  "value3",
		"field6":  3,
		"field7":  "value4",
		"field8":  4,
		"field9":  "value5",
		"field10": 5,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for log15
func BenchmarkLog15_With10Fields(b *testing.B) {
	l := log15.New()
	l.SetHandler(log15.StreamHandler(io.Discard, log15.JsonFormat()))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message",
			"field1", "value1",
			"field2", 1,
			"field3", "value2",
			"field4", 2,
			"field5", "value3",
			"field6", 3,
			"field7", "value4",
			"field8", 4,
			"field9", "value5",
			"field10", 5,
		)
	}
}

// Benchmark for logrus
func BenchmarkLogrus_With10Fields(b *testing.B) {
	l := logrus.New()
	l.SetOutput(io.Discard)
	l.SetFormatter(&logrus.JSONFormatter{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{
			"field1":  "value1",
			"field2":  1,
			"field3":  "value2",
			"field4":  2,
			"field5":  "value3",
			"field6":  3,
			"field7":  "value4",
			"field8":  4,
			"field9":  "value5",
			"field10": 5,
		}).Info("test message")
	}
}

// ============================================================================
// Benchmarks: Log a message with a logger that already has 10 fields of context
// ============================================================================

// Benchmark for this logger with pre-configured context
func BenchmarkLogger_WithContext(b *testing.B) {
	config := logger.Config{
		LogstashEnabled: false,
		AppName:         "bench-app",
		MinLevel:        logger.INFO,
		AsyncEnabled:    false,
	}

	baseLogger := logger.New(config)
	defer baseLogger.Close()
	baseLogger.SetOutput(io.Discard)

	// Create logger with context fields
	l := baseLogger.With(map[string]interface{}{
		"field1":  "value1",
		"field2":  1,
		"field3":  "value2",
		"field4":  2,
		"field5":  "value3",
		"field6":  3,
		"field7":  "value4",
		"field8":  4,
		"field9":  "value5",
		"field10": 5,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for zap with pre-configured context
func BenchmarkZap_WithContext(b *testing.B) {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	l := zap.New(core).With(
		zap.String("field1", "value1"),
		zap.Int("field2", 1),
		zap.String("field3", "value2"),
		zap.Int("field4", 2),
		zap.String("field5", "value3"),
		zap.Int("field6", 3),
		zap.String("field7", "value4"),
		zap.Int("field8", 4),
		zap.String("field9", "value5"),
		zap.Int("field10", 5),
	)
	defer l.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for zap (sugared) with pre-configured context
func BenchmarkZapSugared_WithContext(b *testing.B) {
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	l := zap.New(core).Sugar().With(
		"field1", "value1",
		"field2", 1,
		"field3", "value2",
		"field4", 2,
		"field5", "value3",
		"field6", 3,
		"field7", "value4",
		"field8", 4,
		"field9", "value5",
		"field10", 5,
	)
	defer l.Sync()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for zerolog with pre-configured context
func BenchmarkZerolog_WithContext(b *testing.B) {
	l := zerolog.New(io.Discard).With().
		Str("field1", "value1").
		Int("field2", 1).
		Str("field3", "value2").
		Int("field4", 2).
		Str("field5", "value3").
		Int("field6", 3).
		Str("field7", "value4").
		Int("field8", 4).
		Str("field9", "value5").
		Int("field10", 5).
		Logger()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("test message")
	}
}

// Benchmark for slog with pre-configured context
func BenchmarkSlog_WithContext(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil)).With(
		"field1", "value1",
		"field2", 1,
		"field3", "value2",
		"field4", 2,
		"field5", "value3",
		"field6", 3,
		"field7", "value4",
		"field8", 4,
		"field9", "value5",
		"field10", 5,
	)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for slog (LogAttrs) with pre-configured context
func BenchmarkSlogLogAttrs_WithContext(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil)).With(
		slog.String("field1", "value1"),
		slog.Int("field2", 1),
		slog.String("field3", "value2"),
		slog.Int("field4", 2),
		slog.String("field5", "value3"),
		slog.Int("field6", 3),
		slog.String("field7", "value4"),
		slog.Int("field8", 4),
		slog.String("field9", "value5"),
		slog.Int("field10", 5),
	)

	b.ResetTimer()
	b.ReportAllocs()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		l.LogAttrs(ctx, slog.LevelInfo, "test message")
	}
}

// Benchmark for go-kit with pre-configured context
func BenchmarkGoKit_WithContext(b *testing.B) {
	l := kitlog.NewJSONLogger(io.Discard)
	l = kitlog.With(l,
		"field1", "value1",
		"field2", 1,
		"field3", "value2",
		"field4", 2,
		"field5", "value3",
		"field6", 3,
		"field7", "value4",
		"field8", 4,
		"field9", "value5",
		"field10", 5,
	)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Log("msg", "test message")
	}
}

// Benchmark for log15 with pre-configured context
func BenchmarkLog15_WithContext(b *testing.B) {
	l := log15.New()
	l.SetHandler(log15.StreamHandler(io.Discard, log15.JsonFormat()))
	l = l.New("field1", "value1", "field2", 1, "field3", "value2", "field4", 2, "field5", "value3",
		"field6", 3, "field7", "value4", "field8", 4, "field9", "value5", "field10", 5)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for apex/log with pre-configured context
func BenchmarkApexLog_WithContext(b *testing.B) {
	apexlog.SetHandler(apexloghandler.Default)
	l := apexlog.WithFields(apexlog.Fields{
		"field1":  "value1",
		"field2":  1,
		"field3":  "value2",
		"field4":  2,
		"field5":  "value3",
		"field6":  3,
		"field7":  "value4",
		"field8":  4,
		"field9":  "value5",
		"field10": 5,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("test message")
	}
}

// Benchmark for logrus with pre-configured context
func BenchmarkLogrus_WithContext(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithFields(logrus.Fields{
		"field1":  "value1",
		"field2":  1,
		"field3":  "value2",
		"field4":  2,
		"field5":  "value3",
		"field6":  3,
		"field7":  "value4",
		"field8":  4,
		"field9":  "value5",
		"field10": 5,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		entry.Info("test message")
	}
}
