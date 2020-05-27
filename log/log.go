package log

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Sugar *zap.SugaredLogger
var cfg zap.Config

func init() {
	cfg = zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05Z07"))
	}
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeDuration = nil
	cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(caller.TrimmedPath())
	}
	logger, _ := cfg.Build(zap.AddCallerSkip(1))
	Sugar = logger.Sugar()
}

func Logger(skip int) *zap.Logger {
	logger, _ := cfg.Build(zap.AddCallerSkip(skip))
	return logger
}

func SqlLog(template string, args ...interface{}) {
	Sugar.Infof(strings.ReplaceAll(template, "?", "%v"), args...)
}

func Infof(template string, args ...interface{}) error {
	msg := fmtMsg(template, args...)
	Sugar.Infof(msg)
	return errors.New(msg)
}
func Infow(msg string, keysAndValues ...interface{}) {
	Sugar.Infow(msg, keysAndValues...)
}
func Infoc(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		Sugar.Infow(msg, keysAndValues...)
	}
}

func Warnf(template string, args ...interface{}) error {
	msg := fmtMsg(template, args...)
	Sugar.Warnf(msg)
	return errors.New(msg)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	Sugar.Warnw(msg, keysAndValues...)
}
func Warnc(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		Sugar.Warnw(msg, keysAndValues...)
	}
}

func Errorf(template string, args ...interface{}) error {
	msg := fmtMsg(template, args...)
	Sugar.Errorf(msg)
	return errors.New(msg)
}
func Errorw(msg string, keysAndValues ...interface{}) {
	Sugar.Errorw(msg, keysAndValues...)
}
func Errorc(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		Sugar.Errorw(msg, keysAndValues...)
	}
}

func Fatalf(template string, args ...interface{}) error {
	msg := fmtMsg(template, args...)
	Sugar.Fatalf(msg)
	return errors.New(msg)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	Sugar.Fatalw(msg, keysAndValues...)
}
func Fatalc(err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		Sugar.Fatalw(msg, keysAndValues...)
	}
}

func fmtMsg(template string, args ...interface{}) string {
	if template == "" && len(args) > 0 {
		return fmt.Sprint(args...)
	} else if template != "" && len(args) > 0 {
		return fmt.Sprintf(template, args...)
	} else {
		return template
	}
}
