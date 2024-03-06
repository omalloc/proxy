package proxy

import (
	"context"
	"fmt"
	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/node/direct"
	"github.com/omalloc/proxy/selector/random"
	"net"
	"net/http"
	"time"
)

type ReverseProxy struct {
	// Rebalancer is nodes rebalancer.
	selector.Rebalancer
	*http.Client
	*direct.Builder

	selector selector.Selector
}

type Option func(*ReverseProxy)

func New(opts ...Option) *ReverseProxy {
	r := &ReverseProxy{
		Client:   &http.Client{},
		Builder:  &direct.Builder{},
		selector: random.NewBuilder().Build(),
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *ReverseProxy) Do(req *http.Request) (*http.Response, error) {
	selected, done, err := r.selector.Select(req.Context())
	if err != nil {
		return nil, selector.ErrNoAvailable
	}
	defer done(req.Context(), selector.DoneInfo{
		Err:           err,
		BytesSent:     true,
		BytesReceived: true,
	})

	// Set the URL to the selected TCP address
	r.Client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			fmt.Println("RemoteAddr:", selected.Address())
			return net.Dial("tcp", selected.Address())
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       3 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return r.Client.Do(req)
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

// WithClient is set http.Client
func WithClient(client *http.Client) Option {
	return func(r *ReverseProxy) {
		r.Client = client
	}
}
