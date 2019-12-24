package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/archive/logger"
	"github.com/WabisabiNeet/CollectSuperChat/collector"
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

func main() {
	defer logger.Info("--------------------------------------------------------")

	channel := flag.String("c", "", "channel")
	start := flag.String("start", "", "start date eg. 20191212")
	end := flag.String("end", "", "end date eg. 20191213")

	vid := flag.String("v", "", "video id")
	flag.Parse()

	if *channel == "" && *vid == "" {
		logger.Fatal("required -c or -v option.")
	}
	if *channel != "" && *vid != "" {
		logger.Fatal("-c and -v cannot be used at the same time.")
	}
	if *channel != "" && (*start == "" || *end == "") {
		logger.Fatal("-c option is required start and end option.")
	}

	var Ids []string
	if *channel != "" {
		var err error
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
		Ids = getVideoIDs(ys, *channel, startTime, endTime, "")
	} else if *vid != "" {
		Ids = append(Ids, *vid)
	} else {
		logger.Fatal("invalid state.")
	}
	logger.Info("%v", Ids)

	c := &collector.Collector{
		ID:             "0",
		YoutubeService: ys,
	}

	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)
	wg := sync.WaitGroup{}
	for _, id := range Ids {
		c.StartWatch(&wg, id, true)
		select {
		case <-quit:
			return
		default:
		}
	}
}
