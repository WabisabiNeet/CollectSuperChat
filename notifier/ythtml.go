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

	t := time.NewTicker(30 * time.Minute)
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
		vid, err := livestream.GetLiveIDFromChannelPage(item.Snippet.ResourceId.ChannelId)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		n.CollectChat(vid)
	}
	n.checkAndStartSubscribedChannel(ys, res.NextPageToken)
}
