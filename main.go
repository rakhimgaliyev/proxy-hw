package main

import (
	"log"
	"net/http"

	proxy "github.com/rakhimgaliyev/goproxy/proxy"
)

func main() {
	proxy := proxy.NewProxyHttpServer()
	log.Fatal(http.ListenAndServe(":8080", proxy))
}
