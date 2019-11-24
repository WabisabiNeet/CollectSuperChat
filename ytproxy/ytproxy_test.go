package ytproxy_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	"github.com/elazarl/goproxy"
)

var flagHost = "localhost"

func TestProxy(tt *testing.T) {
	flag.Parse()

	proxy2 := goproxy.NewProxyHttpServer()

	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.MitmConnect, host
	})

	re := regexp.MustCompile(`.*/get_live_chat.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		referer := resp.Request.Header["Referer"]
		fmt.Println(fmt.Sprintf("livestream referer:%v", referer))
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(fmt.Sprintf("body:%v", string(body)))
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	})
	re2 := regexp.MustCompile(`/live_chat_replay.*`)
	proxy2.OnResponse(goproxy.UrlMatches(re2)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		referer := resp.Request.Header["Referer"]
		fmt.Println(fmt.Sprintf("archive referer:%v", referer))
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(fmt.Sprintf("body:%v", string(body)))
		}

		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	})

	proxy2.Verbose = false
	http.ListenAndServe("0.0.0.0:8081", proxy2)
	// cfg, err := ytproxy.TLSConfigFromCA(&goproxy.GoproxyCa, flagHost)
	// if err != nil {
	// 	log.Println(err)
	// }
	// sv2 := &http.Server{ // http1.1+http2 and tls1.2
	// 	Handler:   proxy2,
	// 	Addr:      "127.0.0.1:8083",
	// 	TLSConfig: cfg.Clone(),
	// }
	// sv2.TLSConfig.NextProtos = []string{"http/1.1", "h2"}
	// sv2.TLSConfig.MinVersion = tls.VersionTLS12
	// http2.VerboseLogs = true
	// http2.ConfigureServer(sv2, nil)
	// sv2.ListenAndServeTLS("", "")
}
