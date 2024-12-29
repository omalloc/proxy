package selector

import (
	"context"
)

// staticSelector is composite selector.
type staticSelector struct {
	NodeBuilder WeightedNodeBuilder
	Balancer    Balancer
	node        WeightedNode
}

func donef(ctx context.Context, di DoneInfo) {}

func (d *staticSelector) Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error) {
	if p, ok := FromPeerContext(ctx); ok {
		return p.Node, donef, nil
	}

	return d.node, donef, nil
}

func (d *staticSelector) Apply(nodes []Node) {
	d.node = d.NodeBuilder.Build(nodes[0])
}

// StaticNodeBuilder is de
type StaticNodeBuilder struct {
	Node     WeightedNodeBuilder
	Balancer BalancerBuilder
}

// Build create builder
func (db *StaticNodeBuilder) Build() Selector {
	return &staticSelector{
		NodeBuilder: db.Node,
		Balancer:    db.Balancer.Build(),
	}
}
