package goproxy

import (
	"io"
	"net"
	"net/http"
)

func (proxy *ProxyHttpServer) handleHttps(w http.ResponseWriter, r *http.Request) {
	hij, _ := w.(http.Hijacker)

	proxyClient, _, _ := hij.Hijack()
	host := r.URL.Host
	targetSiteCon, _ := net.Dial("tcp", host)
	proxyClient.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	targetTCP, _ := targetSiteCon.(*net.TCPConn)
	proxyClientTCP, _ := proxyClient.(*net.TCPConn)
	go copyAndClose(targetTCP, proxyClientTCP)
	go copyAndClose(proxyClientTCP, targetTCP)
}

func copyAndClose(dst, src *net.TCPConn) {
	io.Copy(dst, src)
	dst.CloseWrite()
	src.CloseRead()
}
