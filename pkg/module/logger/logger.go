package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger representation
type Logger interface {
	With(args ...any) Logger
	WithOptions(opts ...zap.Option) Logger

	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Panic(args ...any)
	Fatal(args ...any)

	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Panicf(template string, args ...any)
	Fatalf(template string, args ...any)

	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Panicw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

var skipOne = zap.AddCallerSkip(1) //nolint:gochecknoglobals

func Debug(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Debug(args...)
}

func Info(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Info(args...)
}

func Warn(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Warn(args...)
}

func Error(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Error(args...)
}

func Panic(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Panic(args...)
}

func Fatal(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatal(args...)
}

func Debugf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Debugf(template, args...)
}

func Infof(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Infof(template, args...)
}

func Warnf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Warnf(template, args...)
}

func Errorf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Errorf(template, args...)
}

func Panicf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Panicf(template, args...)
}

func Fatalf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatalf(template, args...)
}

func Debugw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Debugw(msg, keysAndValues...)
}

func Infow(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Infow(msg, keysAndValues...)
}

func Warnw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Warnw(msg, keysAndValues...)
}

func Errorw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Errorw(msg, keysAndValues...)
}

func Panicw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Panicw(msg, keysAndValues...)
}

func Fatalw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatalw(msg, keysAndValues...)
}
