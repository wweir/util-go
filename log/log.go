package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Cfg   zap.Config
	sugar *zap.SugaredLogger
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

	SetZapSugar(NewZapLogger(zap.AddCallerSkip(1)))
}

func NewZapLogger(opts ...zap.Option) *zap.Logger {
	logger, _ := Cfg.Build(opts...)
	return logger
}

// Handy log functions
var Infow, Warnw, Errorw, Panicw, Fatalw func(msg string, keysAndValues ...interface{})

// SetZapSugar set a custom logger to be the default logger
// the logger should init with option: zap.AddCallerSkip(1)
func SetZapSugar(logger *zap.Logger) {
	sugar = logger.Sugar()

	var defaultLogger *zapLogger
	Infow = defaultLogger.Infow
	Warnw = defaultLogger.Warnw
	Errorw = defaultLogger.Errorw
	Panicw = defaultLogger.Panicw
	Fatalw = defaultLogger.Fatalw
}

// Err will log the message if err != nil , eg:
// log.Err(err).Errorw("XXX", "err", err)
func Err(err error) *zapLogger {
	return &zapLogger{
		err: &err,
	}
}

// ErrPt will log the message if err != nil , eg:
// defer log.ErrPt(err).Errorw("XXX", "err", err)
func ErrPt(err *error) *zapLogger {
	return &zapLogger{
		err: err,
	}
}

type zapLogger struct {
	err *error
}

func (z *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	if z == nil {
		sugar.Infow(msg, keysAndValues...)
	} else if *z.err != nil {
		sugar.With("err", *z.err).Infow(msg, keysAndValues...)
	}
}
func (z *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	if z == nil {
		sugar.Warnw(msg, keysAndValues...)
	} else if *z.err != nil {
		sugar.With("err", *z.err).Warnw(msg, keysAndValues...)
	}
}
func (z *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	if z == nil {
		sugar.Errorw(msg, keysAndValues...)
	} else if *z.err != nil {
		sugar.With("err", *z.err).Errorw(msg, keysAndValues...)
	}
}
func (z *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	if z == nil {
		sugar.Panicw(msg, keysAndValues...)
	} else if *z.err != nil {
		sugar.With("err", *z.err).Panicw(msg, keysAndValues...)
	}
}
func (z *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	if z == nil {
		sugar.Fatalw(msg, keysAndValues...)
	} else if *z.err != nil {
		sugar.With("err", *z.err).Fatalw(msg, keysAndValues...)
	}
}
