package chromedp

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
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

	resChs      = map[string](chan interface{}){}
	resChsMutex = sync.Mutex{}

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// InitChrome opne chrome process.
func InitChrome() error {
	// c, cancel := chromedp.NewContext(context.Background())
	var opts []chromedp.ExecAllocatorOption
	for _, opt := range chromedp.DefaultExecAllocatorOptions {
		opts = append(opts, opt)
	}
	// no headless
	opts = append(opts,
		// chromedp.Flag("headless", false),
		// chromedp.Flag("hide-scrollbars", false),
		// chromedp.Flag("mute-audio", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.DisableGPU,
	)
	aCtx, aCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	c, cancel := chromedp.NewContext(aCtx)
	err := chromedp.Run(c)
	if err != nil {
		return err
	}
	allocCancel = aCancel
	windowCtx = c
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
func OpenLiveChatWindow(vid string, isArchive bool) (<-chan string, error) {
	tabMutex.Lock()
	defer tabMutex.Unlock()

	_, ok := tabs[vid]
	if ok {
		return nil, errors.New("already started")
	}

	var yturl string
	if isArchive {
		yturl = archiveURL
	} else {
		yturl = liveStreamChatURL
	}
	u, _ := url.Parse(yturl)
	q := u.Query()
	q.Add("v", vid)
	u.RawQuery = q.Encode()

	ctx, cancel := chromedp.NewContext(windowCtx)

	chromedp.ListenTarget(
		ctx,
		func(ev interface{}) {
			fmt.Println(reflect.TypeOf(ev))
			switch ev.(type) {
			case *network.EventResponseReceived:
				ev := ev.(*network.EventResponseReceived)
				fmt.Println(fmt.Sprintf("event received:%v", ev.Response.URL))
				fmt.Println(fmt.Sprintf("event received:%+v", ev))
				// fmt.Println(ev.Type)

				if ev.Type != "XHR" {
					return
				}
				if !strings.Contains(ev.Response.URL, "get_live_chat") && !strings.Contains(ev.Response.URL, "get_live_chat_replay") {
					return
				}

				reqID := ev.RequestID.String()
				var ok bool
				func() {
					resChsMutex.Lock()
					defer resChsMutex.Unlock()
					_, ok = resChs[reqID]
				}()
				if ok {
					return
				}
				ch := make(chan interface{}, 1)
				resChs[reqID] = ch

				go func() {
					<-ch // waiting to LoadingFinished

					c := chromedp.FromContext(ctx)
					rbp := network.GetResponseBody(ev.RequestID)
					body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						log.Info(err.Error())
						return
					}

					err = sendToWatcher(vid, string(body))
					if err != nil {
						log.Warn(err.Error())
					}
				}()
			case *network.EventLoadingFinished:
				ev := ev.(*network.EventLoadingFinished)
				fmt.Println(fmt.Sprintf("event received:%+v", ev))

				resChsMutex.Lock()
				defer resChsMutex.Unlock()
				reqID := ev.RequestID.String()
				ch, ok := resChs[reqID]
				if ok {
					delete(resChs, reqID)
					close(ch)
				}
			}
		},
	)

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(u.String()),
		chromedp.Sleep(time.Second*1),
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

// AlreadyStarted returns bool
func AlreadyStarted(vid string) bool {
	tabMutex.Lock()
	defer tabMutex.Unlock()
	_, ok := tabs[vid]
	return ok
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
