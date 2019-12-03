package ytproxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sync"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/elazarl/goproxy"
)

var watcher = map[string](chan string){}
var watcherMutex = sync.Mutex{}

// OpenYoutubeLiveChatProxy open youtube proxy.
func OpenYoutubeLiveChatProxy() {
	proxy2 := goproxy.NewProxyHttpServer()
	proxy2.Verbose = false

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re := regexp.MustCompile(`www.youtube.com.*/get_live_chat.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(OnLiveChatResponse)
	re2 := regexp.MustCompile(`www.youtube.com/.*get_live_chat_replay.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re2)).DoFunc(OnLiveChatReplayResponse)

	sv2 := &http.Server{
		Handler: proxy2,
		Addr:    "0.0.0.0:8081",
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := sv2.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Error("HTTP server Shutdown: %v", err)
		}
	}()
	go sv2.ListenAndServe()
}

// OnLiveChatResponse is proxy func.
func OnLiveChatResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	referer := resp.Request.Header["Referer"]
	if len(referer) == 0 {
		return resp
	}
	vid := ""
	for _, u := range referer {
		url, _ := url.Parse(u)
		q := url.Query()
		vid = q.Get("v")
	}
	if vid == "" {
		return resp
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return resp
	}

	json := string(body)
	if w, ok := getWatcher(vid); ok {
		w <- json
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp
}

// OnLiveChatReplayResponse is proxy func.
func OnLiveChatReplayResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// リプレイ取得時はRefererにvidが含まれないためチェックをしない

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info(err.Error())
		return resp
	}

	json := string(body)
	for _, w := range watcher {
		w <- json
		break
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp
}

func getWatcher(vid string) (chan<- string, bool) {
	watcherMutex.Lock()
	defer watcherMutex.Lock()

	w, ok := watcher[vid]
	return w, ok
}

// CreateWatcher is register channel
func CreateWatcher(vid string) <-chan string {
	watcherMutex.Lock()
	defer watcherMutex.Lock()

	w, ok := watcher[vid]
	if ok {
		return w
	}

	newCh := make(chan string, 20)
	watcher[vid] = newCh

	return newCh
}

// UnsetWatcher is unregister channel
func UnsetWatcher(vid string) {
	watcherMutex.Lock()
	defer watcherMutex.Lock()

	w, ok := watcher[vid]
	if ok {
		close(w)
	}
	delete(watcher, vid)
}
