package aio

import (
	"context"

	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/node/direct"
)

const (
	// Name is aio(All in One, only one node) balancer name
	Name = "aio"
)

var _ selector.Balancer = (*Balancer)(nil) // Name is balancer name

// Option is aio builder option.
type Option func(o *options)

// options is aio builder options
type options struct{}

// Balancer is a aio balancer.
type Balancer struct {
	node selector.WeightedNode
}

// New aio a selector.
func New(opts ...Option) selector.Selector {
	return NewBuilder(opts...).Build()
}

// Pick is pick a weighted node.
func (p *Balancer) Pick(_ context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	return nil, nil, nil
}

// NewBuilder returns a selector builder with wrr balancer
func NewBuilder(opts ...Option) selector.Builder {
	var option options
	for _, opt := range opts {
		opt(&option)
	}
	return &selector.StaticNodeBuilder{
		Balancer: &Builder{},
		Node:     &direct.Builder{},
	}
}

// Builder is wrr builder
type Builder struct{}

// Build creates Balancer
func (b *Builder) Build() selector.Balancer {
	return &Balancer{}
}
