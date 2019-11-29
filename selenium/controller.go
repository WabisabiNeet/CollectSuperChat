package selenium

import (
	"net/url"
	"sync"

	"github.com/sclevine/agouti"
)

const liveStreamChatURL = "https://www.youtube.com/live_chat?is_popout=1&"

var pages map[string]*agouti.Page
var pagesMutex = sync.Mutex{}

// OpenLiveChatWindow send request to selenium server that opening live chat window
func OpenLiveChatWindow(vid string) error {
	seleniumServer := "http://selenium:4444/wd/hub"
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
		return err
	}

	pagesMutex.Lock()
	defer pagesMutex.Unlock()
	pages[vid] = page

	u, _ := url.Parse(liveStreamChatURL)
	q := u.Query()
	q.Add("v", vid)

	err = page.Navigate(u.String())
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
