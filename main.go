package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	ctx := context.Background()

	apikey := os.Getenv("YOUTUBE_API_KEY")
	if apikey == "" {
		log.Fatal("not found api key.")
	}

	ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		log.Fatal(err)
	}

	// e.g. https://www.youtube.com/watch?v=WziZomD9KC8
	chatid, err := getLiveChatID(ys, "WziZomD9KC8")
	if err != nil {
		log.Fatal(err)
	}

	messages, nextToken, err := getLiveChatMessage(ys, chatid, "")
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Print(messages)

	for nextToken != "" {
		log.Printf("next token:%v\n", nextToken)
		messages, nextToken, err = getLiveChatMessage(ys, chatid, nextToken)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(messages)
	}
}

func getLiveChatID(ys *youtube.Service, vid string) (string, error) {
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

func getLiveChatMessage(ys *youtube.Service, cid, next string) (messages []string, nextToken string, err error) {
	call := ys.LiveChatMessages.List(cid, "snippet,authorDetails")
	call.PageToken(next)
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

	log.Printf("interval: %vms\n", res.PollingIntervalMillis)
	time.Sleep(time.Duration(res.PollingIntervalMillis) * time.Millisecond)
	return
}
