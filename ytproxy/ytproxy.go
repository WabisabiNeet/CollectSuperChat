package ytproxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"sync"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
)

var watcher = map[string](chan string){}
var watcherMutex = sync.Mutex{}

// OpenYoutubeLiveChatProxy open youtube proxy.
func OpenYoutubeLiveChatProxy(port int) int {
	proxy2 := goproxy.NewProxyHttpServer()
	proxy2.Verbose = false

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re := regexp.MustCompile(`www.youtube.com.*/get_live_chat.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(OnLiveChatResponse)

	sv2 := &http.Server{
		Handler: proxy2,
		Addr:    fmt.Sprintf("0.0.0.0:%v", port),
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

	ln, err := net.Listen("tcp", sv2.Addr)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	go sv2.Serve(ln)

	log.Info("addr:%v", ln.Addr().String())
	return ln.Addr().(*net.TCPAddr).Port
}

// OpenYoutubeLiveChatReplayProxy open youtube proxy.
func OpenYoutubeLiveChatReplayProxy(port int) int {
	proxy2 := goproxy.NewProxyHttpServer()
	proxy2.Verbose = false

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re2 := regexp.MustCompile(`www.youtube.com.*/get_live_chat_replay.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re2)).DoFunc(OnLiveChatReplayResponse)

	sv2 := &http.Server{
		Handler: proxy2,
		Addr:    fmt.Sprintf("0.0.0.0:%v", port),
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

	ln, err := net.Listen("tcp", sv2.Addr)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	go sv2.Serve(ln)

	log.Info("addr:%v", ln.Addr().String())
	return ln.Addr().(*net.TCPAddr).Port
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
	err = sendToWatcher(vid, json)
	if err != nil {
		log.Error(errors.Wrapf(err, "vid:[%v] json:[%v]", vid, json).Error())
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
		log.Error(err.Error())
		return resp
	}

	json := string(body)
	err = sendToWatcher("replay", json)
	if err != nil {
		log.Error(errors.Wrapf(err, "vid:[%v] json:[%v]", "replay", json).Error())
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp
}

func sendToWatcher(vid, json string) error {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	w, ok := watcher[vid]
	if !ok {
		return errors.New("watcher not found")
	}
	w <- json
	return nil
}

// CreateWatcher is register channel
func CreateWatcher(vid string) (<-chan string, error) {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	_, ok := watcher[vid]
	if ok {
		return nil, errors.New("already started")
	}

	newCh := make(chan string, 100)
	watcher[vid] = newCh

	return newCh, nil
}

// UnsetWatcher is unregister channel
func UnsetWatcher(vid string) {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	w, ok := watcher[vid]
	if !ok {
		return
	}
	delete(watcher, vid)
	close(w)
}
