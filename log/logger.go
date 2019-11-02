package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var dbglog *zap.Logger

func init() {
	initDebugLogger()
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

// GetLogger return debug logger.
func GetLogger() *zap.Logger {
	return dbglog
}
