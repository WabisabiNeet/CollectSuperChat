package selenium

import (
	"net/url"
	"sync"
	"time"

	"github.com/sclevine/agouti"
)

const (
	liveStreamChatURL = "https://www.youtube.com/live_chat?is_popout=1"
	archiveURL        = "https://www.youtube.com/watch"
)

var pages = map[string]*agouti.Page{}
var pagesMutex = sync.Mutex{}

// OpenLiveChatWindow send request to selenium server that opening live chat window
func OpenLiveChatWindow(vid string) error {
	seleniumServer := "http://selenium:4444/wd/hub"
	options := []agouti.Option{
		agouti.Browser("chrome"),
		agouti.ChromeOptions(
			"args", []string{
				"--proxy-server=collector:8081",
				"--ignore-certificate-errors",
			}),
	}
	// free proxy 43.245.216.189:8080
	page, err := agouti.NewPage(seleniumServer, options...)
	if err != nil {
		return err
	}

	pagesMutex.Lock()
	defer pagesMutex.Unlock()
	pages[vid] = page

	u, _ := url.Parse(liveStreamChatURL)
	q := u.Query()
	q.Add("v", vid)
	u.RawQuery = q.Encode()

	for i := 0; i < 3; i++ {
		err = page.Navigate(u.String())
		if err == nil {
			break
		}
		time.Sleep(time.Second * 3)
	}

	return err
}

// OpenArchiveWindow send request to selenium server that opening live chat window
func OpenArchiveWindow(vid string) error {
	seleniumServer := "http://selenium:4444/wd/hub"
	options := []agouti.Option{
		agouti.Browser("chrome"),
		agouti.ChromeOptions(
			"args", []string{
				"--proxy-server=archive_collector:8082",
				"--ignore-certificate-errors",
				"--autoplay-policy=no-user-gesture-required",
			}),
	}
	// free proxy 43.245.216.189:8080
	page, err := agouti.NewPage(seleniumServer, options...)
	if err != nil {
		return err
	}

	pagesMutex.Lock()
	defer pagesMutex.Unlock()
	pages[vid] = page

	u, _ := url.Parse(archiveURL)
	q := u.Query()
	q.Add("v", vid)
	u.RawQuery = q.Encode()

	for i := 0; i < 3; i++ {
		err = page.Navigate(u.String())
		if err == nil {
			break
		}
		time.Sleep(time.Second * 3)
	}

	return err
}

// CloseLiveChatWindow send request to selenium server that closing live chat window
func CloseLiveChatWindow(vid string) error {
	pagesMutex.Lock()
	defer pagesMutex.Unlock()

	const maxCloseRetry = 3
	page, ok := pages[vid]
	if !ok {
		return nil
	}

	defer page.Session().Delete()
	var err error
	for i := 0; i < maxCloseRetry; i++ {
		err = page.CloseWindow()
		if err == nil {
			delete(pages, vid)
			return nil
		}
	}

	return err
}

// CloseAllLiveChatWindow send request to selenium server that closing live chat window
func CloseAllLiveChatWindow() {
	pagesMutex.Lock()
	defer pagesMutex.Unlock()

	for _, page := range pages {
		page.CloseWindow()
	}
	pages = map[string]*agouti.Page{}
}
