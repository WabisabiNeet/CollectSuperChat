package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var chatlog, dbglog *zap.Logger

func init() {
	initDebugLogger()
}

func initSuperChatLogger(cannelID string) {
	logfolder := path.Join("superchat", cannelID)
	os.MkdirAll(logfolder, os.ModeDir|0755)
	today := time.Now()
	const layout = "200601"
	filename := logfolder + "/" + today.Format(layout) + ".txt"

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
		OutputPaths: []string{"stdout", filename},
		// ErrorOutputPaths: []string{"stderr"},
	}
	chatlog, _ = myConfig.Build()
}

func initDebugLogger() {
	os.MkdirAll("debug", os.ModeDir|0755)
	today := time.Now()
	const layout = "200601"
	filename := "./debug/" + today.Format(layout) + ".txt"

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)

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
		OutputPaths:      []string{"stdout", filename},
		ErrorOutputPaths: []string{"stderr"},
	}
	dbglog, _ = myConfig.Build()
}

func main() {
	channelID := flag.String("c", "", "a channelID")
	flag.Parse()

	if *channelID == "" {
		dbglog.Fatal("channelID is nil")
	}
	initSuperChatLogger(*channelID)

	ctx := context.Background()

	apikey := os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
	if apikey == "" {
		dbglog.Fatal("not found api key.")
	}

	ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		dbglog.Fatal(err.Error())
	}

	vid, err := livestream.GetLiveStreamID(ys, *channelID)
	for err != nil {
		e1 := err.Error()
		if e1 == "live stream not found" {
			dbglog.Info(err.Error())
			time.Sleep(time.Minute)
			vid, err = livestream.GetLiveStreamID(ys, *channelID)
		} else {
			log.Fatal(err)
		}
	}

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	chatid, err := livestream.GetLiveChatID(ys, vid)
	if err != nil {
		dbglog.Fatal(err.Error())
	}

	nextToken := ""
	for {
		messages, nextToken, err := livestream.GetSuperChatRawMessages(ys, chatid, nextToken)
		if err != nil {
			switch t := err.(type) {
			case *googleapi.Error:
				if t.Code == 403 && t.Message == "The live chat is no longer live." {
					dbglog.Info(t.Message)
					return
				}
			default:
				dbglog.Fatal(t.Error())
			}
		}

		for _, message := range messages {
			message.AuthorDetails.ProfileImageUrl = ""
			message.Etag = ""
			message.Id = ""
			message.Kind = ""
			message.Snippet.AuthorChannelId = ""
			message.Snippet.DisplayMessage = ""
			outputJSON, err := json.Marshal(*message)
			if err == nil {
				chatlog.Info(string(outputJSON))
			}
		}

		if nextToken == "" {
			break
		}
	}
}
