package proxy_test

import (
	"fmt"
	"github.com/omalloc/proxy"
	"github.com/omalloc/proxy/selector"
	"net/http"
	"net/http/httputil"
	"testing"
)

func TestProxy(t *testing.T) {
	p := proxy.New()
	nodes := make([]selector.Node, 0, 10)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, p.Build(selector.NewNode("http", fmt.Sprintf("127.0.0.1:82%02d", i), map[string]string{"weight": "10"})))
	}
	p.Apply(nodes)

	req, err := http.NewRequest("GET", "http://example.com/af", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 第一次发起请求
	resp, err := p.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err := httputil.DumpResponse(resp, false)
	fmt.Println(string(b))

	fmt.Println("------------------------------")
	fmt.Println()

	// 重新设置可用上游节点
	p.Apply([]selector.Node{p.Build(selector.NewNode("http", "127.0.0.1:8200", map[string]string{"weight": "100"}))})

	resp, err = p.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err = httputil.DumpResponse(resp, false)
	fmt.Println(string(b))
}
