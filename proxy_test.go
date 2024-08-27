package proxy_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/omalloc/proxy"
	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/wrr"
)

var (
	p = proxy.New()
)

func TestMain(m *testing.M) {
	nodes := make([]selector.Node, 0, 2)
	for i := 0; i < 2; i++ {
		nodes = append(nodes, p.Build(selector.NewNode("http", fmt.Sprintf("127.0.0.1:828%d", i+1), map[string]string{"weight": "10"})))
	}
	p.Apply(nodes)

	m.Run()
}

func TestProxy(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/path/to/1.apk", nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := httputil.DumpRequest(req, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	// 第一次发起请求
	resp, err := p.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err = httputil.DumpResponse(resp, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	fmt.Println("------------------------------")
	fmt.Println()

	// 重新设置可用上游节点
	p.Apply([]selector.Node{p.Build(selector.NewNode("http", "127.0.0.1:8284", map[string]string{"weight": "100"}))})

	resp, err = p.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err = httputil.DumpResponse(resp, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}

func TestProxyRebalancer(t *testing.T) {
	px := proxy.New(
		proxy.WithSelector(wrr.NewBuilder().Build()),
		proxy.WithInitialNodes([]selector.Node{
			selector.NewNode("http", "127.0.0.1:8282", selector.RawMetadata("weight", "10")),
			selector.NewNode("http", "127.0.0.1:8283", selector.RawMetadata("weight", "20")),
			selector.NewNode("http", "127.0.0.1:8284", selector.RawMetadata("weight", "70")),
		}),
	)

	doProxy := func(i int) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "http://example.com/path/to/5.apk", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := px.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("doProxy index [%d] unexpected status code: %d", i, resp.StatusCode)
		}

		_, _ = io.ReadAll(resp.Body)
		t.Logf("doProxy index [%d] proxy-by: %s\n", i, resp.Header.Get("X-Proxy-By"))
	}

	for i := 0; i < 100; i++ {
		doProxy(i)
	}

}
