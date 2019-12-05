package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/WabisabiNeet/CollectSuperChat/selenium"
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
		log.Error("vid is nil")
		return
	}

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	videoInfo, err := livestream.GetLiveInfo(c.YoutubeService, vid)
	if err != nil {
		log.Error(fmt.Sprintf("[vid:%v] %v", vid, err))
		return
	} else if videoInfo.LiveStreamingDetails.ActiveLiveChatId == "" {
		log.Info(fmt.Sprintf("[vid:%v] Live chat not active.", vid))
		return
	}

	chatlog := initSuperChatLogger(videoInfo.Snippet.ChannelId)
	defer chatlog.Sync()

	ch := ytproxy.CreateWatcher(vid)
	defer ytproxy.UnsetWatcher(vid)

	defer selenium.CloseLiveChatWindow(vid)
	err = selenium.OpenLiveChatWindow(vid)
	if err != nil {
		log.Error(fmt.Sprintf("OpenLiveChatWindow error:%v", err.Error()))
		return
	}

	for {
		select {
		case json, ok := <-ch:
			if !ok {
				log.Warn("channel closed. [%v][%v][%v]",
					videoInfo.Snippet.ChannelTitle,
					videoInfo.Snippet.Title,
					fmt.Sprintf("https://www.youtube.com/watch?v=%v", vid))
				return
			}

			messages, finished, err := livestream.GetLiveChatMessagesFromProxy(json)
			if err != nil {
				log.Error(err.Error())
			}
			if finished {
				log.Info("Live end. [%v][%v][%v]",
					videoInfo.Snippet.ChannelTitle,
					videoInfo.Snippet.Title,
					fmt.Sprintf("https://www.youtube.com/watch?v=%v", vid))
				return
			}

			outputSuperChat(messages, videoInfo, chatlog)
		case <-quit:
			return
		}
	}
}

func outputSuperChat(messages []*livestream.ChatMessage, vinfo *youtube.Video, chatlog *zap.Logger) {
	for _, m := range messages {
		m.VideoInfo.ChannelID = vinfo.Snippet.ChannelId
		m.VideoInfo.ChannelTitle = vinfo.Snippet.ChannelTitle
		m.VideoInfo.VideoID = vinfo.Id
		m.VideoInfo.VideoTitle = vinfo.Snippet.Title

		scheduledStartTime, err := time.Parse(time.RFC3339, vinfo.LiveStreamingDetails.ScheduledStartTime)
		if err != nil {
			m.VideoInfo.ScheduledStartTime = scheduledStartTime.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)
		} else {
			m.VideoInfo.ScheduledStartTime = vinfo.LiveStreamingDetails.ScheduledStartTime
		}
		actualStartTime, err := time.Parse(time.RFC3339, vinfo.LiveStreamingDetails.ActualStartTime)
		if err != nil {
			m.VideoInfo.ActualStartTime = actualStartTime.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)
		} else {
			m.VideoInfo.ActualStartTime = vinfo.LiveStreamingDetails.ActualStartTime
		}

		if m.Message.AmountDisplayString != "" {
			err := m.ConvertToJPY()
			if err != nil {
				log.Warn(err.Error())
			}
		}

		outputJSON, err := json.Marshal(m)
		if err != nil {
			log.Error(err.Error())
		}
		chatlog.Info(string(outputJSON))
	}
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
				log.Info(fmt.Sprintf("[%v] repert getLiveStreamID.", channel))
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
		OutputPaths: []string{filename},
		// ErrorOutputPaths: []string{"stderr"},
	}
	chatlog, _ := myConfig.Build()
	return chatlog
}
