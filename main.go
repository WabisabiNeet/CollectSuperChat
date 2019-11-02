package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/notifier"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const interval = 120 // 90sec.

var dbglog *zap.Logger
var apikey string

func init() {
	initDebugLogger()

	// apikey = os.Getenv("YOUTUBE_API_KEY")
	apikey = os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
	// apikey = os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
	if apikey == "" {
		dbglog.Fatal("not found api key.")
	}

	// watching ticket folder.
	os.MkdirAll("waching", os.ModeDir|0755)
}

func initSuperChatLogger(channelID string) *zap.Logger {
	logfolder := path.Join("superchat", channelID)
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
	chatlog, _ := myConfig.Build()
	return chatlog
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
		},
		OutputPaths:      []string{"stdout", filename},
		ErrorOutputPaths: []string{"stderr"},
	}
	dbglog, _ = myConfig.Build()
}

func getChannels() ([]string, error) {
	channels := make([]string, 0)
	flag.Parse()
	filenames := flag.Args()

	f, err := os.Open(filenames[0])
	if err != nil {

		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		channels = append(channels, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return channels, nil
}

func startWatch(wg *sync.WaitGroup, vid string) {
	defer wg.Done()
	if vid == "" {
		dbglog.Fatal("vid is nil")
	}

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	ctx := context.Background()
	ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		dbglog.Warn(err.Error())
		return
	}

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	channel, chatid, err := livestream.GetLiveInfo(ys, vid)
	if err != nil {
		dbglog.Warn(fmt.Sprintf("[%v] %v", channel, err))
		return
	} else if chatid == "" {
		dbglog.Info(fmt.Sprintf("[%v] Live chat not active.", channel))
		return
	}

	chatlog := initSuperChatLogger(channel)
	defer chatlog.Sync()

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
			message.AuthorDetails = nil
			message.Etag = ""
			message.Id = ""
			message.Kind = ""
			message.Snippet.LiveChatId = ""
			outputJSON, err := json.Marshal(*message)
			if err == nil {
				chatlog.Info(string(outputJSON))
			}
		}

		if nextToken == "" {
			dbglog.Info("nextToken is empty.")
			break
		}

		dbglog.Info(fmt.Sprintf("[%v] interval: %vsec.", channel, interval))
		select {
		case <-time.Tick(interval * time.Second):
		case <-quit:
			dbglog.Info(fmt.Sprintf("[%v] signaled.", channel))
			return
		}
	}
}

// func getLiveStreamID(ys *youtube.Service, channel string, sig chan os.Signal) (string, error) {
// 	t := time.NewTicker(time.Minute)
// 	for {
// 		vid, err := livestream.GetLiveStreamID(ys, channel)
// 		if err != nil {
// 			if err.Error() != "live stream not found" {
// 				return "", err
// 			}
// 			select {
// 			case <-t.C:
// 				dbglog.Info(fmt.Sprintf("[%v] repert getLiveStreamID.", channel))
// 				continue
// 			case <-sig:
// 				return "", fmt.Errorf("signaled")
// 			}
// 		}

// 		return vid, nil
// 	}
// }

func main() {
	defer dbglog.Sync()

	wg := &sync.WaitGroup{}
	n := notifier.Gmail{
		CollectChat: func(vid string) {
			wg.Add(1)
			go startWatch(wg, vid)
		},
	}

	n.PollingStart()
	wg.Wait()
}
