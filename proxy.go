package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/node/direct"
	"github.com/omalloc/proxy/selector/random"
)

var DefaultProxy = New()

type Proxy interface {
	Do(req *http.Request) (*http.Response, error)
	Apply(nodes []selector.Node)
}

type ReverseProxy struct {
	// Rebalancer is nodes rebalancer.
	selector.Rebalancer
	*direct.Builder

	dialer    *net.Dialer
	clientMap map[string]*http.Client
	selector  selector.Selector
}

type Option func(*ReverseProxy)

func New(opts ...Option) *ReverseProxy {
	r := &ReverseProxy{
		Builder:   &direct.Builder{},
		clientMap: make(map[string]*http.Client, 16),
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		selector: random.NewBuilder().Build(), // default algorithm is random
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *ReverseProxy) Do(req *http.Request) (*http.Response, error) {
	current, done, err := r.selector.Select(req.Context())
	if err != nil {
		return nil, selector.ErrNoAvailable
	}
	defer done(req.Context(), selector.DoneInfo{
		Err:           err,
		BytesSent:     true,
		BytesReceived: true,
	})

	resp, err := r.find(current.Address()).Do(req)
	resp.Header.Add("X-Proxy-By", fmt.Sprintf("proxy@%s", current.Address()))

	return resp, err
}

func (r *ReverseProxy) find(addr string) *http.Client {
	if client, ok := r.clientMap[addr]; ok {
		return client
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxConnsPerHost:       500,
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return r.dialer.DialContext(ctx, network, addr)
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	r.clientMap[addr] = client

	return client
}

// Apply is apply all nodes when any changes happen
func (r *ReverseProxy) Apply(nodes []selector.Node) {
	r.selector.Apply(nodes)
}

// WithInitialNodes is set initial nodes
func WithInitialNodes(nodes []selector.Node) Option {
	return func(r *ReverseProxy) {
		r.selector.Apply(nodes)
	}
}

// WithSelector is set new-selector
func WithSelector(s selector.Selector) Option {
	return func(r *ReverseProxy) {
		r.selector = s
	}
}

// WithDialer is set custom net.Dialer
func WithDialer(d *net.Dialer) Option {
	return func(r *ReverseProxy) {
		r.dialer = d
	}
}
