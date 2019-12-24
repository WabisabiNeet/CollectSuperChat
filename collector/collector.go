package collector

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/WabisabiNeet/CollectSuperChat/selenium"
	"github.com/WabisabiNeet/CollectSuperChat/ytproxy"
	"google.golang.org/api/youtube/v3"
)

// Collector is service struct
type Collector struct {
	ID              string
	YoutubeService  *youtube.Service
	ProcessingCount int32
}

// StartWatch collect super chat.
func (c *Collector) StartWatch(wg *sync.WaitGroup, vid string, isArchive bool) {
	defer c.DecrementCount()
	defer wg.Done()
	if vid == "" {
		log.Error("vid is nil")
		return
	}

	var watcherID string
	if isArchive {
		watcherID = "replay"
	} else {
		watcherID = vid
	}
	ch, err := ytproxy.CreateWatcher(watcherID)
	if err != nil {
		log.Info(err.Error())
		return
	}
	defer ytproxy.UnsetWatcher(watcherID)

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	videoInfo, err := livestream.GetLiveInfo(c.YoutubeService, vid)
	if err != nil {
		log.Error(fmt.Sprintf("[vid:%v] %v", vid, err))
		return
	} else if !isArchive && videoInfo.LiveStreamingDetails.ActiveLiveChatId == "" {
		log.Info(fmt.Sprintf("[vid:%v] Live chat not active.", vid))
		return
	}

	if isArchive {
		defer selenium.CloseLiveChatWindow(vid)
		err = selenium.OpenArchiveWindow(vid)
	} else {
		defer selenium.CloseLiveChatWindow(vid)
		err = selenium.OpenLiveChatWindow(vid)
	}
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

			var messages []*livestream.ChatMessage
			var finished bool
			if isArchive {
				messages, finished, err = livestream.GetReplayChatMessagesFromProxy(json)
			} else {
				messages, finished, err = livestream.GetLiveChatMessagesFromProxy(json)
			}

			if err != nil {
				log.Error(err.Error())
			}
			if finished {
				log.Info("watch end. [%v][%v][%v]",
					videoInfo.Snippet.ChannelTitle,
					videoInfo.Snippet.Title,
					fmt.Sprintf("https://www.youtube.com/watch?v=%v", vid))
				return
			}

			outputSuperChat(messages, videoInfo, isArchive)
		case <-quit:
			return
		}
	}
}

func outputSuperChat(messages []*livestream.ChatMessage, vinfo *youtube.Video, isArchive bool) {
	for _, m := range messages {
		m.VideoInfo.ChannelID = vinfo.Snippet.ChannelId
		m.VideoInfo.ChannelTitle = vinfo.Snippet.ChannelTitle
		m.VideoInfo.VideoID = vinfo.Id
		m.VideoInfo.VideoTitle = vinfo.Snippet.Title

		scheduledStartTime, err := time.Parse(time.RFC3339, vinfo.LiveStreamingDetails.ScheduledStartTime)
		if err != nil {
			m.VideoInfo.ScheduledStartTime = vinfo.LiveStreamingDetails.ScheduledStartTime
		} else {
			m.VideoInfo.ScheduledStartTime = scheduledStartTime.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)
		}
		actualStartTime, err := time.Parse(time.RFC3339, vinfo.LiveStreamingDetails.ActualStartTime)
		if err != nil {
			m.VideoInfo.ActualStartTime = vinfo.LiveStreamingDetails.ActualStartTime
		} else {
			m.VideoInfo.ActualStartTime = actualStartTime.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339)
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
		o := string(outputJSON)

		if !isArchive {
			err = log.SendChat(vinfo.Snippet.ChannelId, m.Message.MessageID, o)
			if err != nil {
				log.Error(err.Error())
			}
		}

		log.OutputSuperChat(o)
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

func (c *Collector) IncrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), 1)
}

func (c *Collector) DecrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), -1)
}
