package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/WabisabiNeet/CollectSuperChat/notifier"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// MaxKeys is api keys.
const MaxKeys = 9

// Collector is service struct
type Collector struct {
	ID              string
	YoutubeService  *youtube.Service
	ProcessingCount int32
}

// Collectors is List
type Collectors []Collector

var dbglog *zap.Logger
var apikey string
var collectors Collectors

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
		collectors = append(collectors, Collector{
			ID:              string(i),
			YoutubeService:  ys,
			ProcessingCount: 0,
		})
	}
	if len(collectors) == 0 {
		dbglog.Fatal("not found api key.")
	}
}

func initSuperChatLogger(channelID string) *zap.Logger {
	logfolder := path.Join("superchat", channelID)
	os.MkdirAll(logfolder, os.ModeDir|0755)
	today := time.Now()
	const layout = "20060102"
	filename := path.Join(logfolder, fmt.Sprintf("%s.txt", today.Format(layout)))

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.InfoLevel)

	myConfig := zap.Config{
		Level:    level,
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "", // ignore.
			LevelKey:       "", // ignore.
			NameKey:        "Name",
			CallerKey:      "", // ignore.
			MessageKey:     "Msg",
			StacktraceKey:  "St",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths: []string{"stdout", filename},
		// ErrorOutputPaths: []string{"stderr"},
	}
	chatlog, _ := myConfig.Build()
	return chatlog
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

func (c *Collector) incrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), 1)
}

func (c *Collector) decrementCount() {
	c.ProcessingCount = atomic.AddInt32(&(c.ProcessingCount), -1)
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
