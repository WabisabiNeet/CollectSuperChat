package livestream

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/antonholmquist/jason"
	"github.com/pkg/errors"
)

const (
	baseCheckLiveURL = "https://www.youtube.com/channel/%s/videos?view=2&flow=grid"
	selectorBadge    = ".yt-badge-live"
	selectorVideoID  = "h3.yt-lockup-title > a.yt-ui-ellipsis-2"
)

// UpcomingLiveInfo struct
type UpcomingLiveInfo struct {
	StartTime time.Time
	VideoID   string
}

// GetUpcommingLiveIDFromChannelPage return channel id.
func GetUpcommingLiveIDFromChannelPage(channel string) (upcomingInfo []*UpcomingLiveInfo, err error) {
	channelURL := fmt.Sprintf(baseCheckLiveURL, channel)
	u, err := url.Parse(channelURL)
	if err != nil {
		return nil, err
	}
	return GetUpcommingLiveID(u)
}

// GetUpcommingLiveID return channel id.
func GetUpcommingLiveID(u *url.URL) (upcomingInfo []*UpcomingLiveInfo, err error) {
	vhtml, err := getHTMLString(u.String())
	if err != nil {
		return nil, err
	}

	jsonStr, err := getytInitialData(vhtml)
	if err != nil {
		return nil, err
	}

	jsonStr = strings.TrimRight(jsonStr, ";")
	root, err := jason.NewObjectFromReader(strings.NewReader(jsonStr))
	if err != nil {
		return nil, err
	}

	tabs, err := root.GetObjectArray("contents", "twoColumnBrowseResultsRenderer", "tabs")
	if err != nil {
		return nil, err
	}

	for _, tab := range tabs {
		tabURL, err := tab.GetString("tabRenderer", "endpoint", "commandMetadata", "webCommandMetadata", "url")
		if err != nil {
			log.Info(err.Error())
			continue
		}
		if !strings.HasSuffix(tabURL, "videos") {
			continue
		}

		contents, err := tab.GetObjectArray("tabRenderer", "content", "sectionListRenderer", "contents")
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		l := len(contents)
		switch {
		case l == 0:
			continue
		case l > 1:
			log.Warn(fmt.Sprintf("unexpected contents length[%v]", len(contents)))
		}

		contents2, err := contents[0].GetObjectArray("itemSectionRenderer", "contents")
		if err != nil {
			continue
		}
		l = len(contents2)
		switch {
		case l == 0:
			continue
		case l > 1:
			log.Warn(fmt.Sprintf("unexpected contents2 length[%v]", len(contents)))
		}

		items, err := contents2[0].GetObjectArray("gridRenderer", "items")
		if err != nil {
			continue
		}

		for _, item := range items {
			startTime, err := item.GetString("gridVideoRenderer", "upcomingEventData", "startTime")
			if err != nil {
				continue
			}

			s, err := strconv.ParseInt(startTime, 10, 64)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			vid, err := item.GetString("gridVideoRenderer", "videoId")
			if err != nil {
				continue
			}
			upcomingInfo = append(upcomingInfo, &UpcomingLiveInfo{
				StartTime: time.Unix(s, 0),
				VideoID:   vid,
			})
		}
		return upcomingInfo, nil
	}

	return nil, errors.Wrap(errors.New("unexpected error"), jsonStr)
}

func getHTMLString(url string) (string, error) {
	c := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	return string(byteArray), nil
}

func getytInitialData(html string) (string, error) {
	stringReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return "", err
	}

	scripts := doc.Find("script")
	jsonStr := ""
	scripts.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		script := strings.TrimSpace(s.Text())
		if !strings.Contains(script, `window["ytInitialData"]`) {
			return true
		}

		scanner := bufio.NewScanner(strings.NewReader(script))
		for bufSize := bufio.MaxScanTokenSize; !scanner.Scan() && scanner.Err() == bufio.ErrTooLong; bufSize *= 2 {
			if bufSize > 10*1024*1024 {
				return false
			}
			scanner = bufio.NewScanner(strings.NewReader(script))
			scanner.Buffer(make([]byte, bufSize), bufSize)
		}

		ytInitialData := scanner.Text()
		for !strings.Contains(ytInitialData, `window["ytInitialData"]`) && scanner.Scan() {
			ytInitialData = scanner.Text()
		}

		splits := strings.SplitAfter(ytInitialData, ` = `)
		if len(splits) < 2 {
			return true
		}

		jsonStr = splits[1]
		return false
	})

	if jsonStr == "" {
		return "", errors.New("json parse error?")
	}

	return jsonStr, nil
}
