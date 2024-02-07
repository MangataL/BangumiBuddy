package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config logger config
type Config struct {
	Level    zapcore.Level `yaml:"level" json:"level"`       // log level，debug,info,warn,error,fatal
	Filename string        `yaml:"fileName" json:"fileName"` // log file path, absolute path
}

// Build new logger from Config
func (c *Config) Build() (*ZapLogger, error) {
	lv := zap.NewAtomicLevelAt(c.Level)
	var opts []zap.Option

	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(2))
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	file := &lumberjack.Logger{
		Filename:   c.Filename,
		MaxSize:    100, 
		MaxBackups: 1,   
		LocalTime:  true,
	}
	return NewZapLogger(
		zapcore.NewCore(encoder, zapcore.AddSync(file), lv),
		&lv,
		c.Filename,
		opts...,
	), nil
}
