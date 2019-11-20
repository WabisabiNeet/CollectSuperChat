package ytproxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	"github.com/WabisabiNeet/CollectSuperChat/log"
	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

var dbglog *zap.Logger

var flagHost = "localhost"

// OnRecieveLiveChatData is func.
var OnRecieveLiveChatData func(json string)

func init() {
	dbglog = log.GetLogger()
}

// OpenYoutubeLiveChatProxy open youtube proxy.
func OpenYoutubeLiveChatProxy() {
	proxy2 := goproxy.NewProxyHttpServer()

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re := regexp.MustCompile(`.*/get_live_chat.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		referer := resp.Request.Header["Referer"]
		dbglog.Info(fmt.Sprintf("URL:%v referer:%v", resp.Request.URL, referer))
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		json := string(body)
		if err == nil {
			dbglog.Info(fmt.Sprintf("body:%v", json))
		}

		if OnRecieveLiveChatData != nil {
			go OnRecieveLiveChatData(json)
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

	// proxy2.Verbose = false
	// go http.ListenAndServe("0.0.0.0:8081", proxy2)

	cfg, err := TLSConfigFromCA(&goproxy.GoproxyCa, flagHost)
	if err != nil {
		dbglog.Error(err.Error())
	}
	sv2 := &http.Server{ // http1.1+http2 and tls1.2
		Handler:   proxy2,
		Addr:      "127.0.0.1:8083",
		TLSConfig: cfg.Clone(),
	}
	sv2.TLSConfig.NextProtos = []string{"http/1.1", "h2"}
	sv2.TLSConfig.MinVersion = tls.VersionTLS12
	http2.VerboseLogs = true
	http2.ConfigureServer(sv2, nil)

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

	go sv2.ListenAndServeTLS("", "")
}
