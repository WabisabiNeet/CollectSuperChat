package log

import (
	"fmt"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SuperChatLogger struct
type SuperChatLogger struct {
	channelID string

	l *zap.Logger
}

// NewSuperChatLogger return logger for superchat
func NewSuperChatLogger(channelID string) *SuperChatLogger {
	logfolder := path.Join("superchat", channelID)
	os.MkdirAll(logfolder, os.ModeDir|0755)
	today := time.Now()
	const layout = "20060102"
	filename := path.Join(logfolder, fmt.Sprintf("%s.txt", today.Format(layout)))

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.InfoLevel)

	myConfig := zap.Config{
		Level:    level,
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
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
		},
		OutputPaths: []string{filename},
		// ErrorOutputPaths: []string{"stderr"},
	}
	chatlog, _ := myConfig.Build()
	return &SuperChatLogger{
		channelID: channelID,
		l:         chatlog,
	}
}
