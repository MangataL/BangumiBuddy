package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewContext(t *testing.T) {
	ctx := NewContextWithTraceID("id", zap.Field{Key: "field", Type: zapcore.StringType, String: "value"})
	want := logContext{
		traceID: "id",
		fields:  []zap.Field{zap.String("field", "value")},
	}
	get := ctx.Value(logCtxKey)
	assert.Equal(t, want, get)
}

func TestWithFields(t *testing.T) {
	testCases := []struct {
		name           string
		fields         []zap.Field
		ctx            context.Context
		want           logContext
		compareTraceID bool
	}{
		{
			name: "append log context",
			fields: []zap.Field{
				{Key: "aa", Type: zapcore.StringType, String: "bb"},
			},
			ctx: NewContextWithTraceID("id", zap.Field{Key: "field", Type: zapcore.StringType, String: "value"}),
			want: logContext{
				traceID: "id",
				fields: []zap.Field{
					zap.String("field", "value"),
					zap.String("aa", "bb"),
				},
			},
			compareTraceID: true,
		},
		{
			name: "append background",
			fields: []zap.Field{
				{Key: "aa", Type: zapcore.StringType, String: "bb"},
			},
			ctx: context.Background(),
			want: logContext{
				fields: []zap.Field{
					zap.String("aa", "bb"),
				},
			},
			compareTraceID: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := WithFields(tc.ctx, tc.fields...)
			compareLogContext(ctx, t, tc.want, tc.compareTraceID)
		})
	}
}

func compareLogContext(ctx context.Context, t *testing.T, want logContext, compareTraceID bool) {
	get, ok := ctx.Value(logCtxKey).(logContext)
	require.True(t, ok, "Context not have log context")
	if compareTraceID {
		assert.Equal(t, want, get)
		return
	}
	require.NotEmpty(t, get.traceID, "Trace ID can't be empty")
	get.traceID = ""
	want.traceID = ""
	assert.Equal(t, want, get)
}

func TestZapLogger(t *testing.T) {
	writer, log := newTestLog(t)
	ctx := NewContextWithTraceID("traceID")

	log.Debug(ctx, "aaaa")
	log.Info(ctx, "bbb")
	log.Warn(ctx, "ccc")
	log.Error(ctx, "ddd")

	log.Debugf(ctx, "msg: %v", "aaaa")
	log.Infof(ctx, "msg: %v", "bbb")
	log.Warnf(ctx, "msg: %v", "ccc")
	log.Errorf(ctx, "msg: %v", "ddd")

	log.Debug(context.Background(), "aaa")

	log.SetLevel(zap.ErrorLevel)
	log.Debug(context.Background(), "bbb")

	writer.AssertMessages(
		`DEBUG	aaaa	{"traceID": "traceID"}`,
		`INFO	bbb	{"traceID": "traceID"}`,
		`WARN	ccc	{"traceID": "traceID"}`,
		`ERROR	ddd	{"traceID": "traceID"}`,
		`DEBUG	msg: aaaa	{"traceID": "traceID"}`,
		`INFO	msg: bbb	{"traceID": "traceID"}`,
		`WARN	msg: ccc	{"traceID": "traceID"}`,
		`ERROR	msg: ddd	{"traceID": "traceID"}`,
		`DEBUG	aaa`,
	)
}

func newTestLog(t *testing.T) (*testWriter, *ZapLogger) {
	writer := newTestWriter(t)
	enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	lv := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	core := zapcore.NewCore(enc, zapcore.Lock(zapcore.AddSync(writer)), lv)
	log := NewZapLogger(core, &lv, "stdio")
	return writer, log
}

type testWriter struct {
	buf []string
	t   *testing.T
}

func newTestWriter(t *testing.T) *testWriter {
	return &testWriter{t: t}
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	p = bytes.TrimRight(p, "\n")
	w.t.Logf("%s", p)
	allMsg := string(p)
	msg := allMsg[strings.IndexByte(allMsg, '\t')+1:]
	w.buf = append(w.buf, msg)
	return n, nil
}

func (w *testWriter) AssertMessages(msgs ...string) {
	assert.Equal(w.t, msgs, w.buf, "logged messages did not match")
}
