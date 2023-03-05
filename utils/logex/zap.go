package logex

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once

// EncodingConfig represents the needed encoding configuration for logger
type EncodingConfig struct {
	// Format
	Format string
	// LevelEncoder defines how level is encoded (colors, lowercase, etc.)
	LevelEncoder zapcore.LevelEncoder // lowercase
}

var levelEncoder zapcore.LevelEncoder

// LevelEncoder takes a raw string (as []byte) and returnsthe corresponding zapcore.LevelEncoder
func LevelEncoder(raw []byte) zapcore.LevelEncoder {
	if err := levelEncoder.UnmarshalText(raw); err != nil {
		return nil
	}
	return levelEncoder
}

func defaultEncodingConfig(ec *EncodingConfig) *EncodingConfig {
	if ec == nil {
		ec = &EncodingConfig{}
	}
	if len(ec.Format) == 0 {
		ec.Format = "console"
	}
	if ec.LevelEncoder == nil {
		ec.LevelEncoder = zapcore.CapitalColorLevelEncoder
	}
	return ec
}

func TestLogger(t *testing.T) *zap.Logger {
	return Build(t.Name(), zap.DebugLevel, nil)
}

// Build builds the default zap logger, and sets the global zap logger to the configured logger instance.
func Build(appName string, level zapcore.Level, ec *EncodingConfig) *zap.Logger {
	ec = defaultEncodingConfig(ec)
	cfg := zap.Config{
		Encoding:    ec.Format,
		Level:       zap.NewAtomicLevelAt(level),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: ec.LevelEncoder,
			TimeKey:     "time",
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(iso3339CleanTime(t))
			},
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("err making logger: %+v", err)
	}

	// HACK: callers of Build don't know if it has been called/they don't know if Build will set the zap global logger
	// which means Build has indeterminate behavior
	once.Do(func() {
		zap.ReplaceGlobals(logger)
	})

	return logger
}

// GetLoggerLevelValue resolves logger level to zap level
func GetLoggerLevelValue(loggerLevel string) (zapcore.Level, error) {
	switch loggerLevel {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level - %s", loggerLevel)
	}
}

// iso3339CleanTime converts the given time to ISO 3339 format
func iso3339CleanTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}

// Reset the once init for logger
func Reset() {
	once = sync.Once{}
}
