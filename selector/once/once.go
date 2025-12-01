package once

import (
	"context"

	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/node/direct"
)

const (
	// Name is once(only one node) balancer name
	Name = "once"
)

var _ selector.Balancer = (*Balancer)(nil) // Name is balancer name

// Option is once builder option.
type Option func(o *options)

// options is once builder options
type options struct{}

// Balancer is a once balancer.
type Balancer struct {
}

// New once a selector.
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
