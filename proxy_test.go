package proxy

import (
	"crypto/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/omalloc/proxy/selector"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "with custom dialer",
			opts: []Option{
				WithDialer(&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.opts...)
			assert.NotNil(t, p)
			assert.NotNil(t, p.selector)
			assert.NotNil(t, p.dialer)
			assert.NotNil(t, p.clientMap)
		})
	}
}

func TestReverseProxy_Do(t *testing.T) {
	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// 创建测试节点
	node := &mockNode{scheme: "http", addr: ts.URL[7:]} // 移除 "http://" 前缀

	p := New()
	p.Apply([]selector.Node{node})

	req, err := http.NewRequest("GET", ts.URL, nil)
	assert.NoError(t, err)

	resp, err := p.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReverseProxy_Apply(t *testing.T) {
	p := New()
	nodes := []selector.Node{
		&mockNode{scheme: "http", addr: "localhost:8080"},
		&mockNode{scheme: "http", addr: "localhost:8081"},
	}

	p.Apply(nodes)

	// 验证节点是否被正确应用
	client := p.find("localhost:8080")
	assert.NotNil(t, client)
}

func TestWithOptions(t *testing.T) {
	mockActivate := func(client *http.Client) {}

	tests := []struct {
		name string
		opt  Option
	}{
		{
			name: "WithInitialNodes",
			opt:  WithInitialNodes([]selector.Node{&mockNode{scheme: "http", addr: "localhost:8080"}}),
		},
		{
			name: "WithActivateMock",
			opt:  WithActivateMock(mockActivate),
		},
		{
			name: "WithTransport",
			opt:  WithTransport(&http.Transport{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.opt)
			assert.NotNil(t, p)
		})
	}
}

func TestMockRequest(t *testing.T) {
	// no node proxy
	proxyClient := New(WithActivateMock(httpmock.ActivateNonDefault))
	// mock 1.apk
	httpmock.RegisterResponder(http.MethodGet, "http://example.com/path/to/1.apk", func(r *http.Request) (*http.Response, error) {
		buf := make([]byte, 2<<10)
		_, err := rand.Read(buf)
		return httpmock.NewBytesResponse(http.StatusOK, buf), err
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/path/to/1.apk", nil)

	resp, err := proxyClient.Do(req)
	if err == nil {
		t.Fatal(selector.ErrNoAvailable)
	}

	assert.Equal(t, err, selector.ErrNoAvailable)
	assert.Nil(t, resp)

	// add node
	proxyClient.Apply([]selector.Node{&mockNode{scheme: "http", addr: "localhost:8888"}})

	resp, err = proxyClient.Do(req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	md, _ := httputil.DumpResponse(resp, false)
	t.Logf("all response info: %v", string(md))
}

func TestReUseClient(t *testing.T) {
	proxyClient := New(
		WithInitialNodes([]selector.Node{&mockNode{"http", "127.0.0.1:8888"}}),
		WithActivateMock(httpmock.ActivateNonDefault),
	)
	// mock 1.apk
	httpmock.RegisterResponder(http.MethodGet, "http://example.com/path/to/1.apk", func(r *http.Request) (*http.Response, error) {
		buf := make([]byte, 2<<10)
		_, err := rand.Read(buf)
		return httpmock.NewBytesResponse(http.StatusOK, buf), err
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/path/to/1.apk", nil)

	resp, err := proxyClient.Do(req)

	assert.NotEqual(t, err, selector.ErrNoAvailable)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// re-use client
	resp, err = proxyClient.Do(req)
	assert.NotEqual(t, err, selector.ErrNoAvailable)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestRedirectLocation(t *testing.T) {
	proxyClient := New(
		WithInitialNodes([]selector.Node{&mockNode{"http", "127.0.0.1:8888"}}),
		WithActivateMock(httpmock.ActivateNonDefault),
	)

	// mock 1.apk
	httpmock.RegisterResponder(http.MethodGet, "http://example.com/path/to/1.apk", func(r *http.Request) (*http.Response, error) {
		buf := make([]byte, 2<<10)
		_, err := rand.Read(buf)
		return httpmock.NewBytesResponse(http.StatusOK, buf), err
	})

	redirectCount := 0
	// mock 301
	httpmock.RegisterResponder(http.MethodGet, "http://example.com/path/to/301", func(r *http.Request) (*http.Response, error) {
		redirectCount++

		if redirectCount > 10 {
			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header: http.Header{
					"Location": []string{"http://example.com/path/to/1.apk"},
				},
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusMovedPermanently,
			Header: http.Header{
				"Location": []string{"http://example.com/path/to/301"},
			},
		}, nil
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/path/to/1.apk", nil)

	resp, err := proxyClient.Do(req)
	assert.Nil(t, err)
	assert.NotEqual(t, err, selector.ErrNoAvailable)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

// mockNode 实现 selector.Node 接口
type mockNode struct {
	scheme string
	addr   string
}

func (n *mockNode) Address() string {
	return n.addr
}

func (n *mockNode) Scheme() string {
	return n.scheme
}

func (n *mockNode) Weight() float64 {
	return 1.0
}

func (n *mockNode) InitialWeight() *int64 {
	return nil
}

func (n *mockNode) Version() string {
	return "v1.0.0"
}

func (n *mockNode) Metadata() map[string]string {
	return nil
}
