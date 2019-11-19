// Copyright (C) 2019 RICOH Co., Ltd. All rights reserved.

package ytproxy

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/http2"
)

var flagHost = flag.String("host", "localhost", "")

func TestProxy(tt *testing.T) {
	flag.Parse()

	proxy2 := goproxy.NewProxyHttpServer()
	// re := regexp.MustCompile(`.*/get_live_chat/.*`)
	proxy2.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		fmt.Println("HandleConnectFunc")
		return goproxy.MitmConnect, host
	})
	proxy2.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		fmt.Println("OnRequest().DoFunc")
		// return req, goproxy.NewResponse(req,
		// 	goproxy.ContentTypeText, http.StatusForbidden, "就業時間中です。")
		return req, nil
	})
	proxy2.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		fmt.Println("OnResponse().DoFunc")
		return resp
	})
	// proxy2.OnResponse(goproxy.UrlMatches(re)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// 	fmt.Println("test")
	// 	return resp
	// })
	// proxy2.OnResponse(goproxy.UrlIs(`/search`)).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// 	fmt.Println("ReqHostIs [www.google.co.jp]")
	// 	return resp
	// })
	proxy2.Verbose = true
	go http.ListenAndServe("0.0.0.0:8081", proxy2)

	cfg, err := TLSConfigFromCA(&goproxy.GoproxyCa, *flagHost)
	if err != nil {
		log.Println(err)
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
	sv2.ListenAndServeTLS("", "")
}
