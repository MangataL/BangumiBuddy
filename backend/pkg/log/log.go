// Package log 日志库
package log

import (
	"context"

	"go.uber.org/zap/zapcore"
)

var defaultLogger = NewDefaultZapLogger()

// SetLogger set global logger
func SetLogger(logger *ZapLogger) {
	defaultLogger = logger
}

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Print.
func Debug(ctx context.Context, args ...interface{}) {
	defaultLogger.Debug(ctx, args...)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Debugf(ctx, format, args...)
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func Info(ctx context.Context, args ...interface{}) {
	defaultLogger.Info(ctx, args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Infof(ctx, format, args...)
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func Warn(ctx context.Context, args ...interface{}) {
	defaultLogger.Warn(ctx, args...)
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Warnf(ctx, format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func Error(ctx context.Context, args ...interface{}) {
	defaultLogger.Error(ctx, args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Errorf(ctx, format, args...)
}

// Fatal logs to FATAL log. Arguments are handled in the manner of fmt.Print.
func Fatal(ctx context.Context, args ...interface{}) {
	defaultLogger.Fatal(ctx, args...)
}

// Fatalf logs to FATAL log. Arguments are handled in the manner of fmt.Printf.
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Fatalf(ctx, format, args...)
}

// SetLevel 设置日志级别
func SetLevel(level zapcore.Level) {
	defaultLogger.SetLevel(level)
}

// GetLevel 获取当前日志打印级别
func GetLevel() zapcore.Level {
	return defaultLogger.GetLevel()
}

// GetFileName 获取日志文件名
func GetFileName() string {
	return defaultLogger.GetFileName()
}
