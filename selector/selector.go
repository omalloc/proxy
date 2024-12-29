package selector

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
)

var _ Rebalancer = (*defaultSelector)(nil)

// ErrNoAvailable is no available node.
var ErrNoAvailable = errors.New("no_available_node")

// Selector is node pick balancer.
type Selector interface {
	Rebalancer

	// Select nodes
	// if err == nil, selected and done must not be empty.
	Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error)
}

// Rebalancer is nodes rebalancer.
type Rebalancer interface {
	// Apply is apply all nodes when any changes happen
	Apply(nodes []Node)
}

// Builder build selector
type Builder interface {
	Build() Selector
}

// defaultSelector is composite selector.
type defaultSelector struct {
	NodeBuilder WeightedNodeBuilder
	Balancer    Balancer

	nodes atomic.Value
}

func (d *defaultSelector) Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error) {
	// 快速继承上下文中的 peer 直接执行返回
	if p, ok := FromPeerContext(ctx); ok {
		return p.Node, donef, nil
	}

	// 正常执行选节点
	var (
		options    SelectOptions
		candidates []WeightedNode
	)

	nodes, ok := d.nodes.Load().([]WeightedNode)
	if !ok {
		return nil, nil, ErrNoAvailable
	}

	for _, o := range opts {
		o(&options)
	}
	if len(options.NodeFilters) > 0 {
		newNodes := make([]Node, len(nodes))
		for i, wc := range nodes {
			newNodes[i] = wc
		}
		for _, filter := range options.NodeFilters {
			newNodes = filter(ctx, newNodes)
		}
		candidates = make([]WeightedNode, len(newNodes))
		for i, n := range newNodes {
			candidates[i] = n.(WeightedNode)
		}
	} else {
		candidates = nodes
	}

	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailable
	}

	wn, done, err := d.Balancer.Pick(ctx, candidates)
	if err != nil {
		return nil, nil, err
	}

	p, ok := FromPeerContext(ctx)
	if ok {
		p.Node = wn.Raw()
	}

	return wn.Raw(), done, nil
}

func (d *defaultSelector) Apply(nodes []Node) {
	weightedNodes := make([]WeightedNode, 0, len(nodes))
	for _, n := range nodes {
		weightedNodes = append(weightedNodes, d.NodeBuilder.Build(n))
	}
	// TODO: Do not delete unchanged nodes
	d.nodes.Store(weightedNodes)
}

// DefaultBuilder is de
type DefaultBuilder struct {
	Node     WeightedNodeBuilder
	Balancer BalancerBuilder
}

// Build create builder
func (db *DefaultBuilder) Build() Selector {
	return &defaultSelector{
		NodeBuilder: db.Node,
		Balancer:    db.Balancer.Build(),
	}
}

var _ Node = (*DefaultNode)(nil)

// DefaultNode is selector node
type DefaultNode struct {
	scheme   string
	addr     string
	weight   *int64
	version  string
	name     string
	metadata map[string]string
}

// Scheme is node scheme
func (n *DefaultNode) Scheme() string {
	return n.scheme
}

// Address is node address
func (n *DefaultNode) Address() string {
	return n.addr
}

// ServiceName is node serviceName
func (n *DefaultNode) ServiceName() string {
	return n.name
}

// InitialWeight is node initialWeight
func (n *DefaultNode) InitialWeight() *int64 {
	return n.weight
}

// Version is node version
func (n *DefaultNode) Version() string {
	return n.version
}

// Metadata is node metadata
func (n *DefaultNode) Metadata() map[string]string {
	return n.metadata
}

// NewNode new node
func NewNode(scheme, addr string, metadata map[string]string) Node {
	n := &DefaultNode{
		scheme: scheme,
		addr:   addr,
	}
	if metadata != nil && len(metadata) > 0 {
		n.metadata = metadata
		n.name = metadata["name"]
		if str, ok := metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				n.weight = &weight
			}
		}
	}
	return n
}

func RawMetadata(keyvals ...string) map[string]string {
	l := len(keyvals)
	if l%2 != 0 {
		return map[string]string{}
	}

	metadata := make(map[string]string)
	for i := 0; i < l; i += 2 {
		metadata[keyvals[i]] = keyvals[i+1]
	}
	return metadata
}
