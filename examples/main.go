package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/omalloc/proxy"
	"github.com/omalloc/proxy/selector"
)

func main() {
	// use default proxy client
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy.Apply([]selector.Node{
		selector.NewNode("http", ts.Config.Addr, nil),
	})

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	resp, err := proxy.Do(req)
	if err != nil {
		panic(err)
	}

	info, _ := httputil.DumpResponse(resp, false)
	log.Printf("Response Info : \n%v", string(info))
}
