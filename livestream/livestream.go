package livestream

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/youtube/v3"
)

const maxResult = 2000

var log = logrus.New()

func init() {
	log.Out = os.Stdout
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

// GetLiveChatID return active live chat id.
func GetLiveChatID(ys *youtube.Service, vid string) (string, error) {
	call := ys.Videos.List("LiveStreamingDetails").Id(vid)
	res, err := call.Do()

	if err != nil {
		return "", err

	}

	for _, item := range res.Items {
		return item.LiveStreamingDetails.ActiveLiveChatId, nil
	}

	return "", fmt.Errorf("active chat can not found")
}

// GetSuperChatRawMessages return live chat messages
func GetSuperChatRawMessages(ys *youtube.Service, cid, next string) (messages []*youtube.LiveChatMessage, nextToken string, err error) {
	call := ys.LiveChatMessages.List(cid, "snippet,authorDetails")
	call.PageToken(next)
	call.MaxResults(maxResult)
	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	nextToken = res.NextPageToken
	for _, item := range res.Items {
		switch item.Snippet.Type {
		case "superChatEvent", "superStickerEvent":

		default:
			continue
		}
		messages = append(messages, item)
	}
	return
}
