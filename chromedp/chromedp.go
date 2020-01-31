package chromedp

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	liveStreamChatURL = "https://www.youtube.com/live_chat?is_popout=1"
	archiveURL        = "https://www.youtube.com/watch"
)

type tabInfo struct {
	watcher chan string

	tabCancel context.CancelFunc
}

var (
	allocCancel  context.CancelFunc
	windowCtx    context.Context
	windowCancel context.CancelFunc

	tabs     = map[string](*tabInfo){}
	tabMutex = sync.Mutex{}

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// InitChrome opne chrome process.
func InitChrome() error {
	// ctx, cancel := chromedp.NewContext(context.Background())
	// defer cancel()
	var opts []chromedp.ExecAllocatorOption
	for _, opt := range chromedp.DefaultExecAllocatorOptions {
		opts = append(opts, opt)
	}
	// no headless
	opts = append(opts,
		chromedp.Flag("headless", false),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", true),
	)
	aCtx, aCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	c, cancel := chromedp.NewContext(aCtx)
	err := chromedp.Run(c)
	if err != nil {
		return err
	}

	windowCtx = c
	allocCancel = aCancel
	windowCancel = cancel

	return nil
}

// TerminateChrome close chrome window
func TerminateChrome() {
	if windowCancel != nil {
		windowCancel()
	}
	if allocCancel != nil {
		allocCancel()
	}
}

// OpenLiveChatWindow open live chat window.
func OpenLiveChatWindow(vid string) (<-chan string, error) {
	tabMutex.Lock()
	defer tabMutex.Unlock()

	_, ok := tabs[vid]
	if ok {
		return nil, errors.New("already started")
	}

	u, _ := url.Parse(liveStreamChatURL)
	q := u.Query()
	q.Add("v", vid)
	u.RawQuery = q.Encode()

	ctx, cancel := chromedp.NewContext(windowCtx)

	chromedp.ListenTarget(
		ctx,
		func(ev interface{}) {
			// fmt.Println(reflect.TypeOf(ev))
			if ev, ok := ev.(*network.EventResponseReceived); ok {
				// fmt.Println(fmt.Sprintf("event received:%v", ev.Response.URL))
				// fmt.Println(ev.Type)

				if ev.Type != "XHR" {
					return
				}
				if !strings.Contains(ev.Response.URL, "get_live_chat") {
					return
				}

				go func() {
					// print response body
					c := chromedp.FromContext(windowCtx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(windowCtx, c.Target))
					if err != nil {
						log.Warn(err.Error())
						return
					}

					err = sendToWatcher(vid, string(body))
					if err != nil {
						log.Warn(err.Error())
					}
				}()

			}
		},
	)

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(u.String()),
		chromedp.Sleep(time.Second*5),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	tab := tabInfo{
		tabCancel: cancel,
		watcher:   make(chan string, 100),
	}
	tabs[vid] = &tab
	return tab.watcher, nil
}

// CloseLinveChatWindow close live chat window.
func CloseLinveChatWindow(vid string) {
	tabMutex.Lock()
	defer tabMutex.Unlock()

	tab, ok := tabs[vid]
	if !ok {
		return
	}

	delete(tabs, vid)
	tab.tabCancel()
	close(tab.watcher)
}

func sendToWatcher(vid, json string) error {
	tabMutex.Lock()
	defer tabMutex.Unlock()

	tab, ok := tabs[vid]
	if !ok {
		return errors.New("watcher not found")
	}
	tab.watcher <- json
	return nil
}
