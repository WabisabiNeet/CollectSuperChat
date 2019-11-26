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

// ILogger is original logger interface.
type ILogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Sync()
}

var dbglog *zap.Logger

// Logger is original logger.
type Logger struct {
	ILogger
	dbglog *zap.Logger
}

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
		TimeKey:        "time", // ignore.
		LevelKey:       "",     // ignore.
		NameKey:        "Name",
		CallerKey:      "", // ignore.
		MessageKey:     "Msg",
		StacktraceKey:  "St",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)
	dbglog = zap.New(
		zapcore.NewCore(enc, sink, level),
	)
}

// GetLogger return debug logger(type zap).
func GetLogger() *zap.Logger {
	return dbglog
}

// GetOriginLogger return debug logger(type original).
func GetOriginLogger() ILogger {
	return &Logger{
		dbglog: dbglog,
	}
}

// Debug output debug level.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.dbglog.Debug(fmt.Sprintf(format, args...))
}

// Info output info level.
func (l *Logger) Info(format string, args ...interface{}) {
	l.dbglog.Info(fmt.Sprintf(format, args...))
}

// Warn output warn level.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.dbglog.Warn(fmt.Sprintf(format, args...))
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
}

// Error output error level.
func (l *Logger) Error(format string, args ...interface{}) {
	l.dbglog.Error(fmt.Sprintf(format, args...))
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
}

// Fatal outpu fatal level.
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.dbglog.Fatal(fmt.Sprintf(format, args...))
	sentry.CaptureException(errors.New(fmt.Sprintf(format, args...)))
}

// Sync is wapper: zap.Logger
func (l *Logger) Sync() {
	l.dbglog.Sync()
}
