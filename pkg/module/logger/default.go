package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger Logger

// SetDefaultLogger sets default logger
func SetDefaultLogger(log Logger) {
	defaultLogger = log
}

var atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)

// AtomicLevel returns global log level
func AtomicLevel() zap.AtomicLevel {
	return atomicLevel
}

// SetLevel sets global log level
func SetLevel(lvl string) {
	l := zapcore.DebugLevel
	switch strings.ToLower(lvl) {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	case "warn":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	case "panic":
		l = zapcore.PanicLevel
	case "fatal":
		l = zapcore.FatalLevel
	default:
		// Fallthrough with default level
	}
	atomicLevel.SetLevel(l)
}

func init() {
	var cores []zapcore.Core

	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.MessageKey = "message"
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	cfg.ConsoleSeparator = " "

	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.Lock(zapcore.AddSync(os.Stdout)),
		atomicLevel,
	)
	cores = append(cores, consoleCore)

	log := zap.New(zapcore.NewSamplerWithOptions(
		zapcore.NewTee(cores...), time.Second, 100, 100,
	), zap.AddCaller())
	zap.RedirectStdLog(log)
	zap.ReplaceGlobals(log)
	defaultLogger = &ZapLogger{
		SugaredLogger: log.Sugar(),
	}
}

type ZapLogger struct {
	*zap.SugaredLogger
}

func (l *ZapLogger) With(args ...any) Logger {
	return &ZapLogger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

func (l *ZapLogger) WithOptions(opts ...zap.Option) Logger {
	return &ZapLogger{
		SugaredLogger: l.SugaredLogger.WithOptions(opts...),
	}
}
