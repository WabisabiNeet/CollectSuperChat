package notifier

import (
	"os"
	"os/signal"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
)

// YoutubeHTML struct.
type YoutubeHTML struct {
	CollectChat func(vid string)
}

// PollingStart polling gmail.
func (n *YoutubeHTML) PollingStart() {
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	b := GetCredentials()
	// If modifying these scopes, delete your previously saved token.json.
	config := GetConfig(b)
	client := GetClient(config)

	ys, err := youtube.New(client)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Unable to retrieve Youtube client").Error())
	}

	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		n.checkAndStartSubscribedChannel(ys, "")

		select {
		case <-t.C:
		case <-quit:
			return
		}
	}

}

// CheckAndStartSubscribedChannel rrr
func (n *YoutubeHTML) checkAndStartSubscribedChannel(ys *youtube.Service, nextPageToken string) {
	sub := ys.Subscriptions.List("snippet")
	sub.Mine(true)
	sub.MaxResults(50)
	sub.PageToken(nextPageToken)

	res, err := sub.Do()
	if err != nil {
		log.Error(err.Error())
	}

	for _, item := range res.Items {
		log.Info("[%v] Channel page check start", item.Snippet.ResourceId.ChannelId)

		lives, err := livestream.GetUpcommingLiveIDFromChannelPage(item.Snippet.ResourceId.ChannelId)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		for _, live := range lives {
			now := time.Now()
			if now.Before(live.StartTime) && live.StartTime.Sub(now) < (time.Hour*6) {
				n.CollectChat(live.VideoID)
			}
		}
		time.Sleep(10 * time.Second)
	}

	if res.NextPageToken != "" {
		n.checkAndStartSubscribedChannel(ys, res.NextPageToken)
	}
}
