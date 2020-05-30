package log

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Cfg zap.Config

func init() {
	Cfg = zap.NewProductionConfig()
	Cfg.Encoding = "console"
	Cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("01-02 15:04:05.000Z07"))
	}
	Cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	Cfg.EncoderConfig.EncodeDuration = nil
	Cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(caller.TrimmedPath())
	}
}

func Logger(opts ...zap.Option) *zap.Logger {
	logger, _ := Cfg.Build(opts...)
	return logger
}

type logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

type nopLogger struct{}

func (*nopLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (*nopLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (*nopLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (*nopLogger) Fatalw(msg string, keysAndValues ...interface{}) {}

var (
	checkOnce  sync.Once
	checkSugar *zap.SugaredLogger
	checkInit  = func() {
		checkSugar = Logger().Sugar()
	}
)

// Check will log the message if err not nil, eg:
// defer log.Check(err).Wranw("XXX", "err", err)
func Check(err error) logger {
	checkOnce.Do(checkInit)

	if err == nil {
		return &nopLogger{}
	}

	return checkSugar
}

var (
	defaultOnce  sync.Once
	defaultSugar *zap.SugaredLogger
	defaultInit  = func() {
		defaultSugar = Logger(zap.AddCallerSkip(1)).Sugar()
	}
)

func Infow(msg string, keysAndValues ...interface{}) {
	defaultOnce.Do(defaultInit)
	defaultSugar.Infow(msg, keysAndValues...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	defaultOnce.Do(defaultInit)
	defaultSugar.Warnw(msg, keysAndValues...)
}
func Errorw(msg string, keysAndValues ...interface{}) {
	defaultOnce.Do(defaultInit)
	defaultSugar.Errorw(msg, keysAndValues...)
}
func Panicw(msg string, keysAndValues ...interface{}) {
	defaultOnce.Do(defaultInit)
	defaultSugar.Panicw(msg, keysAndValues...)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	defaultOnce.Do(defaultInit)
	defaultSugar.Fatalw(msg, keysAndValues...)
}
