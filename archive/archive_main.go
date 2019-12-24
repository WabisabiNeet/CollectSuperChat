package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/archive/logger"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var ys *youtube.Service

func init() {
	apikey := os.Getenv("YOUTUBE_WATCH_LIVE_KEY_ARCHIVE")
	// apikey = os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
	if apikey == "" {
		logger.Fatal("not found api key.")
	}

	ctx := context.Background()
	var err error
	ys, err = youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func main() {
	channel := flag.String("c", "", "channel")
	start := flag.String("start", "", "start date eg. 20191212")
	end := flag.String("end", "", "end date eg. 20191213")
	flag.Parse()

	startTime, err := time.Parse("20060102", *start)
	if err != nil {
		logger.Error(err.Error())
		flag.Usage()
		return
	}
	endTime, err := time.Parse("20060102", *end)
	if err != nil {
		logger.Error(err.Error())
		flag.Usage()
		return
	}

	logger.Info("%v %v %v", *channel, startTime, endTime)
	Ids := getVideoIDs(ys, *channel, startTime, endTime, "")
	logger.Info("%v", Ids)
}

func getVideoIDs(ys *youtube.Service, channel string, start, end time.Time, next string) (Ids []string) {
	call := ys.Search.List("id")
	call.ChannelId(channel)
	call.PublishedAfter(start.Format(time.RFC3339))
	call.PublishedBefore(end.Format(time.RFC3339))
	call.MaxResults(50)
	call.PageToken(next)
	call.Type("video")
	call.EventType("completed")

	res, err := call.Do()
	if err != nil {
		return nil
	}

	for _, i := range res.Items {
		Ids = append(Ids, i.Id.VideoId)
	}

	if res.NextPageToken != "" {
		ss := getVideoIDs(ys, channel, start, end, res.NextPageToken)
		Ids = append(Ids, ss...)
	}

	return Ids
}
