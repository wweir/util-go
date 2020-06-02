package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Cfg zap.Config
var defaultSugar *zap.SugaredLogger

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

	defaultSugar = Logger(zap.AddCallerSkip(1)).Sugar()
}

func Logger(opts ...zap.Option) *zap.Logger {
	logger, _ := Cfg.Build(opts...)
	return logger
}

func Infow(msg string, keysAndValues ...interface{}) {
	defaultSugar.Infow(msg, keysAndValues...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Warnw(msg, keysAndValues...)
}
func Errorw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Errorw(msg, keysAndValues...)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Fatalw(msg, keysAndValues...)
}

// If will log the message if should log is true, eg:
// defer log.If(errors.As(err, xxx)).Errorw("XXX", "err", err)
func If(ok bool) logger {
	if ok {
		return &zapLogger{}
	}
	return &nopLogger{}
}

type logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	If(bool) logger
}

type nopLogger struct{}

func (*nopLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (*nopLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (*nopLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (*nopLogger) Fatalw(msg string, keysAndValues ...interface{}) {}
func (*nopLogger) If(bool) logger                                  { return &nopLogger{} }

type zapLogger struct{}

func (*zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	defaultSugar.Infow(msg, keysAndValues...)
}
func (*zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Warnw(msg, keysAndValues...)
}
func (*zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Errorw(msg, keysAndValues...)
}
func (*zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	defaultSugar.Fatalw(msg, keysAndValues...)
}
func (*zapLogger) If(ok bool) logger {
	if ok {
		return &zapLogger{}
	}
	return &nopLogger{}
}
