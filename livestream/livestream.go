package livestream

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/youtube/v3"
)

const interval = 1 * 60 * 1000
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

// GetSuperChatMessage return live chat messages
func GetSuperChatMessages(ys *youtube.Service, cid, next string) (messages []string, nextToken string, err error) {
	call := ys.LiveChatMessages.List(cid, "snippet,authorDetails")
	call.PageToken(next)
	call.MaxResults(2000)
	res, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	nextToken = res.NextPageToken
	var sb strings.Builder
	for _, item := range res.Items {
		sb.Reset()

		sb.WriteString(item.Snippet.PublishedAt)
		sb.WriteString(",")
		sb.WriteString(item.AuthorDetails.DisplayName)
		sb.WriteString(",")
		sb.WriteString(item.AuthorDetails.ChannelId)
		sb.WriteString(",")
		switch item.Snippet.Type {
		case "superChatEvent":
			sb.WriteString(item.Snippet.SuperChatDetails.AmountDisplayString)
			sb.WriteString(",")
			sb.WriteString(fmt.Sprintf("%v", item.Snippet.SuperChatDetails.AmountMicros))
			sb.WriteString(",")
			sb.WriteString(item.Snippet.SuperChatDetails.Currency)
		case "superStickerEvent":
			sb.WriteString(item.Snippet.SuperStickerDetails.AmountDisplayString)
			sb.WriteString(",")
			sb.WriteString(fmt.Sprintf("%v", item.Snippet.SuperStickerDetails.AmountMicros))
			sb.WriteString(",")
			sb.WriteString(item.Snippet.SuperStickerDetails.Currency)
		default:
			continue
		}
		messages = append(messages, sb.String())
	}

	log.Infof("interval: %vms", interval)
	time.Sleep(time.Duration(interval) * time.Millisecond)
	return
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

	log.Infof("interval: %vms", interval)
	time.Sleep(time.Duration(interval) * time.Millisecond)
	return
}
