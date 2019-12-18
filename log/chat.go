package log

import (
	"os"
	"path"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var chatlog *zap.Logger

func init() {
	initSuperChatLogger()
	if ce := chatlog.Check(zap.InfoLevel, "init"); ce == nil {
		Fatal("super chat logger init failed.")
	}
}

// InitSuperChatLogger return chat logger.
func initSuperChatLogger() {
	logfolder := "superchat"
	os.MkdirAll(logfolder, os.ModeDir|0755)
	filename := path.Join(logfolder, "superchat.txt")

	sink := zapcore.AddSync(
		&lumberjack.Logger{
			Filename: filename,
			MaxSize:  10,
		},
	)
	enc := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "", // ignore.
		LevelKey:       "", // ignore.
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
	chatlog = zap.New(
		zapcore.NewCore(enc, sink, level),
	)
}

// OutputSuperChat output chat log
func OutputSuperChat(o string) {
	chatlog.Info(o)
}

// SyncSuerChat is wapper: zap.Logger
func SyncSuerChat() {
	chatlog.Sync()
}
