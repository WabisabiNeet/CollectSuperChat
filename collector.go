package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/ytproxy"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/youtube/v3"
)

// Collector is service struct
type Collector struct {
	ID              string
	YoutubeService  *youtube.Service
	ProcessingCount int32
}

// StartWatch collect super chat.
func (c *Collector) StartWatch(wg *sync.WaitGroup, vid string) {
	defer c.decrementCount()
	defer wg.Done()
	if vid == "" {
		dbglog.Fatal("vid is nil")
	}

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	videoInfo, err := livestream.GetLiveInfo(c.YoutubeService, vid)
	if err != nil {
		dbglog.Warn(fmt.Sprintf("[vid:%v] %v", vid, err))
		return
	} else if videoInfo.LiveStreamingDetails.ActiveLiveChatId == "" {
		dbglog.Info(fmt.Sprintf("[vid:%v] Live chat not active.", vid))
		return
	}

	chatlog := initSuperChatLogger(videoInfo.Snippet.ChannelId)
	defer chatlog.Sync()

	ch := make(chan string, 20)
	ytproxy.SetWatcher(vid, ch)

	// for {
	// 	select {
	// 	case json, ok := <-ch:
	// 	case <-quit:
	// 		return
	// 	}
	// }
}

func getLiveStreamID(ys *youtube.Service, channel string, sig chan os.Signal) (string, error) {
	t := time.NewTicker(time.Minute)
	for {
		vid, err := livestream.GetLiveStreamID(ys, channel)
		if err != nil {
			if err.Error() != "live stream not found" {
				return "", err
			}
			select {
			case <-t.C:
				dbglog.Info(fmt.Sprintf("[%v] repert getLiveStreamID.", channel))
				continue
			case <-sig:
				return "", fmt.Errorf("signaled")
			}
		}

		return vid, nil
	}
}

func (c *Collector) incrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), 1)
}

func (c *Collector) decrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), -1)
}

func initSuperChatLogger(channelID string) *zap.Logger {
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
		OutputPaths: []string{"stdout", filename},
		// ErrorOutputPaths: []string{"stderr"},
	}
	chatlog, _ := myConfig.Build()
	return chatlog
}
