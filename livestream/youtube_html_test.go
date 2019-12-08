package livestream_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"github.com/sclevine/agouti"
)

func TestScrapeLiveStreamingChat(tt *testing.T) {
	vid, err := livestream.GetLiveIDFromChannelPage("UCl_gCybOJRIgOXw6Qb4qJzQ")
	if err != nil {
		tt.Fatal(err)
	}
	fmt.Println(vid)
}

func Test1(tt *testing.T) {
	const baseCheckLiveURL = "https://www.youtube.com/channel/UCQ0UDLQCjY0rmuxCDE38FGg/videos?live_view=501&flow=grid&view=2"

	seleniumServer := "http://192.168.10.11:4444/wd/hub"
	options := []agouti.Option{
		agouti.Browser("chrome"),
		agouti.ChromeOptions(
			"args", []string{
				"--proxy-server=collector:8081",
			}),
	}
	// free proxy 43.245.216.189:8080
	page, err := agouti.NewPage(seleniumServer, options...)
	if err != nil {
		tt.Fatal(err)
	}

	html, err := page.HTML()
	if err != nil {
		tt.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		tt.Fatal(err)
	}

}
