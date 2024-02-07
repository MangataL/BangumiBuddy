package log

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logContextKey string

const (
	logCtxKey logContextKey = "log-context"

	traceID string = "traceID"
)

// ZapLogger 封装zap实现的日志组件
type ZapLogger struct {
	logger   *zap.Logger
	level    *zap.AtomicLevel
	fileName string
}

type logContext struct {
	traceID string
	fields  []zap.Field
}

// ZapLoggerOptions is the option for initializing zap logger
type ZapLoggerOptions func(log *ZapLogger)

// NewContext return logger context with random traceID
func NewContext(fields ...zap.Field) context.Context {
	return NewContextWithTraceID(uuid.New().String(), fields...)
}

func NewContextWithTraceID(traceID string, fields ...zap.Field) context.Context {
	logCtx := logContext{
		traceID: traceID,
		fields:  fields,
	}
	ctx := context.WithValue(context.Background(), logCtxKey, logCtx)
	return ctx
}

// WithFields append fields to context
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	switch logCtx := ctx.Value(logCtxKey).(type) {
	case logContext:
		logCtx.fields = append(logCtx.fields, fields...)
		return context.WithValue(ctx, logCtxKey, logCtx)
	default:
		logCtx = logContext{
			traceID: uuid.New().String(),
			fields:  fields,
		}
		return context.WithValue(ctx, logCtxKey, logCtx)
	}
}

// NewZapLogger 生成ZapLogger
func NewZapLogger(core zapcore.Core, level *zap.AtomicLevel, fileName string, opts ...zap.Option) *ZapLogger {
	zl := &ZapLogger{
		logger:   zap.New(core, opts...),
		level:    level,
		fileName: fileName,
	}
	return zl
}

func NewDefaultZapLogger() *ZapLogger {
	lv := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), lv)
	logger := NewZapLogger(core, &lv, "stdio")
	return logger
}

// Debug logs to DEBUG log, Arguments are handled in the manner of fmt.Print
func (log *ZapLogger) Debug(ctx context.Context, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.DebugLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Debug(fmt.Sprint(args...), traceField...)
	}
}

func (log *ZapLogger) getZapFields(ctx context.Context) []zap.Field {
	logCtx, ok := ctx.Value(logCtxKey).(logContext)
	if !ok {
		return []zap.Field{zap.Skip()}
	}
	traceField := zap.String(traceID, logCtx.traceID)
	ret := []zap.Field{traceField}
	return append(ret, logCtx.fields...)
}

// Debugf logs to DEBUG log, Arguments are handled in the manner of fmt.Printf
func (log *ZapLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.DebugLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Debug(fmt.Sprintf(format, args...), traceField...)
	}
}

// Info logs to INFO log, Arguments are handled in the manner of fmt.Print
func (log *ZapLogger) Info(ctx context.Context, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.InfoLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Info(fmt.Sprint(args...), traceField...)
	}
}

// Infof logs to INFO log, Arguments are handled in the manner of fmt.Printf
func (log *ZapLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.InfoLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Info(fmt.Sprintf(format, args...), traceField...)
	}
}

// Warn logs to WARNING log, Arguments are handled in the manner of fmt.Print
func (log *ZapLogger) Warn(ctx context.Context, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.WarnLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Warn(fmt.Sprint(args...), traceField...)
	}
}

// Warnf logs to WARNING log, Arguments are handled in the manner of fmt.Printf
func (log *ZapLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.WarnLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Warn(fmt.Sprintf(format, args...), traceField...)
	}
}

// Error logs to ERROR log, Arguments are handled in the manner of fmt.Print
func (log *ZapLogger) Error(ctx context.Context, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.ErrorLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Error(fmt.Sprint(args...), traceField...)
	}
}

// Errorf logs to ERROR log, Arguments are handled in the manner of fmt.Printf
func (log *ZapLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.ErrorLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Error(fmt.Sprintf(format, args...), traceField...)
	}
}

// Fatal logs to FATAL log, Arguments are handled in the manner of fmt.Print
func (log *ZapLogger) Fatal(ctx context.Context, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.FatalLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Fatal(fmt.Sprint(args...), traceField...)
	}
}

// Fatalf logs to FATAL log, Arguments are handled in the manner of fmt.Printf
func (log *ZapLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	if log.logger.Core().Enabled(zapcore.FatalLevel) {
		traceField := log.getZapFields(ctx)
		log.logger.Fatal(fmt.Sprintf(format, args...), traceField...)
	}
}

func (log *ZapLogger) GetLevel() zapcore.Level {
	return log.level.Level()
}

// SetLevel set log level
func (log *ZapLogger) SetLevel(level zapcore.Level) {
	log.level.SetLevel(level)
}

// GetFileName 获取日志文件名
func (log *ZapLogger) GetFileName() string {
	return log.fileName
}
