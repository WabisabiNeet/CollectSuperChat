package log

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var dbglog *zap.Logger

func init() {
	initDebugLogger()
	initSentry()
}

func initDebugLogger() {
	os.MkdirAll("debug", os.ModeDir|0755)
	today := time.Now()
	const layout = "200601"
	filename := "./debug/" + today.Format(layout) + ".txt"

	sink := zapcore.AddSync(
		&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    1,
			MaxBackups: 20,
			MaxAge:     20,
		},
	)
	enc := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",  // ignore.
		LevelKey:       "Level", // ignore.
		NameKey:        "Name",
		CallerKey:      "", // ignore.
		MessageKey:     "Msg",
		StacktraceKey:  "St",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     JSTTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)
	dbglog = zap.New(
		zapcore.NewCore(enc, sink, level),
	)
}

// Debug output debug level.
func Debug(format string, args ...interface{}) {
	dbglog.Debug(fmt.Sprintf(format, args...))
}

// Info output info level.
func Info(format string, args ...interface{}) {
	dbglog.Info(fmt.Sprintf(format, args...))
}

// Warn output warn level.
func Warn(format string, args ...interface{}) {
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
	dbglog.Warn(fmt.Sprintf(format, args...))
}

// Error output error level.
func Error(format string, args ...interface{}) {
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
	dbglog.Error(fmt.Sprintf(format, args...))
}

// Fatal outpu fatal level.
func Fatal(format string, args ...interface{}) {
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
	dbglog.Fatal(fmt.Sprintf(format, args...))
}

// Sync is wapper: zap.Logger
func Sync() {
	dbglog.Sync()
	sentry.Flush(time.Second * 10)
}

// JSTTimeEncoder return JST
func JSTTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	enc.AppendString(t.In(jst).Format(time.RFC3339))
}
