package ytproxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
)

var dbglog *zap.Logger

var flagHost = "localhost"

// OnRecieveLiveChatData is func.
var OnRecieveLiveChatData func(vid, json string)

func init() {
	dbglog = log.GetLogger()
}

// OpenYoutubeLiveChatProxy open youtube proxy.
func OpenYoutubeLiveChatProxy() {
	proxy2 := goproxy.NewProxyHttpServer()
	proxy2.Verbose = false

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re := regexp.MustCompile(`.*/get_live_chat.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
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
		dbglog.Info(fmt.Sprintf("URL:%v referer:%v", resp.Request.URL, referer))

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		json := string(body)
		if err == nil {
			dbglog.Info(fmt.Sprintf("body:%v", json))
		}

		if OnRecieveLiveChatData != nil {
			go OnRecieveLiveChatData(vid, json)
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	})
	// re2 := regexp.MustCompile(`/live_chat_replay.*`)
	// proxy2.OnResponse(goproxy.UrlMatches(re2)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// 	referer := resp.Request.Header["Referer"]
	// 	fmt.Println(fmt.Sprintf("archive referer:%v", referer))
	// 	defer resp.Body.Close()
	// 	body, err := ioutil.ReadAll(resp.Body)
	// 	if err == nil {
	// 		fmt.Println(fmt.Sprintf("body:%v", string(body)))
	// 	}

	// 	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// 	return resp
	// })

	sv2 := &http.Server{
		Handler: proxy2,
		Addr:    "127.0.0.1:8081",
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := sv2.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			dbglog.Error(fmt.Sprintf("HTTP server Shutdown: %v", err))
		}
	}()
	go sv2.ListenAndServe()
}
