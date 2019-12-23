package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
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

var collectors = []*Collector{}

func init() {
	for i := 1; i < MaxKeys; i++ {
		apikey := os.Getenv(fmt.Sprintf("YOUTUBE_WATCH_LIVE_KEY%v", i))
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
			ID:              strconv.FormatInt(int64(i), 10),
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

		t := time.NewTicker(12 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info("pollCurrency timer ticked.")
			case <-quit:
				return
			}

			collect()
		}
	}()
}

func main() {
	defer log.Sync()
	defer log.SyncSuerChat()

	m := sync.Mutex{}
	wg := &sync.WaitGroup{}

	f := func(vid string) {
		m.Lock()
		defer m.Unlock()
		sort.Slice(collectors, func(i, j int) bool {
			return collectors[i].ProcessingCount < collectors[j].ProcessingCount
		})
		wg.Add(1)
		collectors[0].incrementCount()
		log.Info(fmt.Sprintf("watch start ID[%v] ProcessingCount[%v]", collectors[0].ID, collectors[0].ProcessingCount))

		go collectors[0].StartWatch(wg, vid, false)
	}

	var ns []notifier.Notifier
	ns = append(ns, &notifier.Gmail{
		CollectChat: f,
	})
	ns = append(ns, &notifier.YoutubeHTML{
		CollectChat: f,
	})

	pollCurrency()
	ytproxy.OpenYoutubeLiveChatProxy()
	for _, n := range ns {
		wg.Add(1)
		go n.PollingStart(wg)
	}
	wg.Wait()
}
