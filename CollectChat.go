package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/currency"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/WabisabiNeet/CollectSuperChat/notifier"
	"github.com/WabisabiNeet/CollectSuperChat/ytproxy"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// MaxKeys is api keys.
const MaxKeys = 9

var apikey string
var collectors []*Collector

func init() {
	for i := 1; i < MaxKeys; i++ {
		apikey = os.Getenv(fmt.Sprintf("YOUTUBE_WATCH_LIVE_KEY%v", i))
		// apikey = os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
		if apikey == "" {
			break
		}

		ctx := context.Background()
		ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
		if err != nil {
			log.Fatal(err.Error())
		}
		collectors = append(collectors, &Collector{
			ID:              string(i),
			YoutubeService:  ys,
			ProcessingCount: 0,
		})
	}
	if len(collectors) == 0 {
		log.Fatal("not found api key.")
	}
}

func pollCurrency() {
	collect := func() {
		for _, c := range currency.Currencies {
			err := c.ScrapeRataToJPY()
			if err != nil {
				log.Warn(err.Error())
			}
			log.Info("[%v] %v", c.Code, c.RateToJPY)
		}
	}

	collect()

	go func() {
		quit := make(chan os.Signal)
		defer close(quit)
		signal.Notify(quit, os.Interrupt)

		for {
			select {
			case <-time.Tick(12 * time.Hour):
			case <-quit:
				return
			}

			collect()
		}
	}()
}

func main() {
	defer log.Sync()

	m := sync.Mutex{}
	wg := &sync.WaitGroup{}
	n := notifier.Gmail{
		CollectChat: func(vid string) {
			m.Lock()
			defer m.Unlock()
			sort.Slice(collectors, func(i, j int) bool {
				return collectors[i].ProcessingCount < collectors[j].ProcessingCount
			})
			wg.Add(1)
			collectors[0].incrementCount()
			log.Info(fmt.Sprintf("watch start ID[%v] ProcessingCount[%v]", collectors[0].ID, collectors[0].ProcessingCount))

			go collectors[0].StartWatch(wg, vid)
		},
	}

	pollCurrency()
	ytproxy.OpenYoutubeLiveChatProxy()
	n.PollingStart()
	wg.Wait()
}
