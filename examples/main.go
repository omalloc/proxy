package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"

	"github.com/omalloc/proxy"
	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/aio"
)

func main() {
	// use default proxy client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	u, _ := url.Parse(ts.URL)

	proxyClient := proxy.New(
		proxy.WithSelector(aio.New()),
		proxy.WithInitialNodes([]selector.Node{
			selector.NewNode(u.Scheme, u.Host, nil),
		}),
	)

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req = req.WithContext(selector.NewPeerContext(req.Context(), &selector.Peer{selector.NewNode("http", "127.0.0.1:8282", nil)}))

	resp, err := proxyClient.Do(req)
	if err != nil {
		panic(err)
	}
	info, _ := httputil.DumpResponse(resp, false)
	log.Printf("Response Info : \n%v", string(info))
}
