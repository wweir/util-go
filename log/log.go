package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Cfg           zap.Config
	defaultLogger Logger

	Infow, Warnw, Errorw, Panicw, Fatalw func(msg string, keysAndValues ...interface{})
	// If will log the message if should log is true, eg:
	// defer log.If(err != nil).Errorw("XXX", "err", err)
	If  func(bool) Logger
	Run func(func()) Logger
)

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

	SetDefaultLogger(&zapLogger{NewZapLogger(zap.AddCallerSkip(1)).Sugar()})
}

func NewZapLogger(opts ...zap.Option) *zap.Logger {
	logger, _ := Cfg.Build(opts...)
	return logger
}

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger

	Infow = defaultLogger.Infow
	Warnw = defaultLogger.Warnw
	Errorw = defaultLogger.Errorw
	Panicw = defaultLogger.Panicw
	Fatalw = defaultLogger.Fatalw
	If = defaultLogger.If
	Run = defaultLogger.Run
}

type Logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	If(bool) Logger
	Run(func()) Logger
}

type NopLogger struct{}

func (*NopLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (*NopLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (*NopLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (*NopLogger) Panicw(msg string, keysAndValues ...interface{}) {}
func (*NopLogger) Fatalw(msg string, keysAndValues ...interface{}) {}
func (*NopLogger) If(bool) Logger                                  { return &NopLogger{} }
func (*NopLogger) Run(func()) Logger                               { return &NopLogger{} }

type zapLogger struct {
	*zap.SugaredLogger
}

func (z *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Infow(msg, keysAndValues...)
}
func (z *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Warnw(msg, keysAndValues...)
}
func (z *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Errorw(msg, keysAndValues...)
}
func (z *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Panicw(msg, keysAndValues...)
}
func (z *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Fatalw(msg, keysAndValues...)
}
func (z *zapLogger) If(ok bool) Logger {
	if ok {
		return z
	}
	return &NopLogger{}
}
func (z *zapLogger) Run(fn func()) Logger {
	fn()
	return z
}
