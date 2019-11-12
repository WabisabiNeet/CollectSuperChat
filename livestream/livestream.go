package livestream

import (
	"fmt"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
)

const maxResult = 2000

var dbglog *zap.Logger

func init() {
	dbglog = log.GetLogger()
}

// GetLiveStreamID return active live ID in specified channnel.
func GetLiveStreamID(ys *youtube.Service, channnelID string) (string, error) {
	call := ys.Search.List("id")
	call.ChannelId(channnelID)
	call.Type("video")
	call.EventType("live")

	res, err := call.Do()
	if err != nil {
		return "", err
	}

	for _, i := range res.Items {
		return i.Id.VideoId, nil
	}

	return "", fmt.Errorf("live stream not found")
}

// GetLiveInfo return video info.
func GetLiveInfo(ys *youtube.Service, vid string) (videoInfo *youtube.Video, err error) {
	call := ys.Videos.List("snippet,LiveStreamingDetails").Id(vid)
	res, err := call.Do()

	if err != nil {
		return nil, err

	}

	for _, item := range res.Items {
		return item, nil
	}

	return nil, fmt.Errorf("active chat can not found")
}

// GetSuperChatRawMessages return live chat messages
func GetSuperChatRawMessages(ys *youtube.Service, cid, next string) (messages []*youtube.LiveChatMessage, nextToken string, intervalMillis int64, err error) {
	dbglog.Info("GetSuperChatRawMessages call.")
	call := ys.LiveChatMessages.List(cid, "snippet,authorDetails")
	call.PageToken(next)
	call.MaxResults(maxResult)
	res, err := call.Do()
	if err != nil {
		return nil, "", 0, err
	}

	intervalMillis = res.PollingIntervalMillis
	nextToken = res.NextPageToken
	for _, item := range res.Items {
		switch item.Snippet.Type {
		case "superChatEvent", "superStickerEvent", "textMessageEvent":
		default:
			continue
		}
		messages = append(messages, item)
	}
	return
}

type ChatMessage struct {
	ChannelID          string `json:"channelId,omitempty"`
	ChannelTitle       string `json:"channelTitle,omitempty"`
	VideoID            string `json:"videoId,omitempty"`
	VideoTitle         string `json:"videotitle,omitempty"`
	ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	ActualStartTime    string `json:"actualStartTime,omitempty"`

	Message *youtube.LiveChatMessage `json:"chat,omitempty"`
}
