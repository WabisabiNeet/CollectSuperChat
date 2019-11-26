package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/WabisabiNeet/CollectSuperChat/notifier"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// MaxKeys is api keys.
const MaxKeys = 9

var dbglog *zap.Logger
var apikey string
var collectors []*Collector

func init() {
	dbglog = log.GetLogger()

	for i := 1; i < MaxKeys; i++ {
		apikey = os.Getenv(fmt.Sprintf("YOUTUBE_WATCH_LIVE_KEY%v", i))
		// apikey = os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
		if apikey == "" {
			break
		}

		ctx := context.Background()
		ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
		if err != nil {
			dbglog.Fatal(err.Error())
		}
		collectors = append(collectors, &Collector{
			ID:              string(i),
			YoutubeService:  ys,
			ProcessingCount: 0,
		})
	}
	if len(collectors) == 0 {
		dbglog.Fatal("not found api key.")
	}
}

func getChannels() ([]string, error) {
	channels := make([]string, 0)
	flag.Parse()
	filenames := flag.Args()

	f, err := os.Open(filenames[0])
	if err != nil {

		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		channels = append(channels, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return channels, nil
}

func main() {
	defer dbglog.Sync()

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
			dbglog.Info(fmt.Sprintf("watch start ID[%v] ProcessingCount[%v]", collectors[0].ID, collectors[0].ProcessingCount))

			go collectors[0].StartWatch(wg, vid)
		},
	}

	n.PollingStart()
	wg.Wait()
}
