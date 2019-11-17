package main

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func TestGetLiveStreamID(tt *testing.T) {
	ctx := context.Background()
	ys, err := youtube.NewService(ctx, option.WithAPIKey(""))
	if err != nil {
		dbglog.Fatal(err.Error())
	}
	c := Collector{
		ID:              "1",
		YoutubeService:  ys,
		ProcessingCount: 0,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	c.StartWatch(wg, "MzgvKiKYGHM")
}
