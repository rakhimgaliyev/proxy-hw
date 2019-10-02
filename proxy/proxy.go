package goproxy

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync/atomic"
)

type ProxyCtx struct {
	Req     *http.Request
	Session int64
	proxy   *ProxyHttpServer
}

type ProxyHttpServer struct {
	sess                   int64
	KeepDestinationHeaders bool
	NonproxyHandler        http.Handler
	Tr                     *http.Transport
	ConnectDial            func(network string, addr string) (net.Conn, error)
}

func copyHeaders(dst, src http.Header, keepDestHeaders bool) {
	if !keepDestHeaders {
		for k := range dst {
			dst.Del(k)
		}
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func (proxy *ProxyHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		proxy.handleHttps(w, r)
	} else {
		ctx := &ProxyCtx{Req: r, Session: atomic.AddInt64(&proxy.sess, 1), proxy: proxy}

		if !r.URL.IsAbs() {
			proxy.NonproxyHandler.ServeHTTP(w, r)
			return
		}

		resp, _ := ctx.proxy.Tr.RoundTrip(r)

		origBody := resp.Body
		defer origBody.Close()

		copyHeaders(w.Header(), resp.Header, proxy.KeepDestinationHeaders)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()
	}
}

func NewProxyHttpServer() *ProxyHttpServer {
	proxy := ProxyHttpServer{
		Tr: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyFromEnvironment},
		NonproxyHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", 500)
		}),
	}
	return &proxy
}
