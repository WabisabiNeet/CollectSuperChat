package livestream

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

const (
	// baseCheckLiveURL = "https://www.youtube.com/channel/__channelID__/videos?live_view=501&flow=grid&view=2"
	baseCheckLiveURL = "https://www.youtube.com/channel/%s/videos?live_view=501&flow=grid&view=2"
	selectorBadge    = ".yt-badge-live"
	selectorVideoID  = "h3.yt-lockup-title > a.yt-ui-ellipsis-2"
)

// GetLiveIDFromChannelPage return channel id.
func GetLiveIDFromChannelPage(channel string) (string, error) {
	// channelURL := strings.Replace(baseCheckLiveURL, "__channelID__", channel, -1)
	channelURL := fmt.Sprintf(baseCheckLiveURL, channel)

	//get
	doc, err := goquery.NewDocument(channelURL)
	if err != nil {
		return "", err
	}

	//check
	isLive := doc.Find(selectorBadge).Size()
	if isLive == 0 {
		return "", fmt.Errorf("Unable to find live")
	}

	videoLink, ok := doc.Find(selectorVideoID).Attr("href")
	if ok != true {
		return "", fmt.Errorf("Unable to find video id element")
	}
	u, err := url.Parse(videoLink)
	if err != nil {
		return "", err
	}
	vid := u.Query().Get("v")
	return vid, nil
}
