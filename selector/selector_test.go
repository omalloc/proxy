package selector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/node/direct"
	"github.com/omalloc/proxy/selector/once"
	"github.com/omalloc/proxy/selector/p2c"
	"github.com/omalloc/proxy/selector/random"
)

func TestSelectNodeWithRandom(t *testing.T) {
	b := direct.Builder{}
	def := random.NewBuilder().Build()

	// 配置3个节点
	nodes := make([]selector.Node, 0, 3)
	nodes = append(nodes, b.Build(selector.NewNode("http", "127.0.0.1:8280", nil)))
	nodes = append(nodes, b.Build(selector.NewNode("http", "127.0.0.1:8281", nil)))
	nodes = append(nodes, b.Build(selector.NewNode("http", "127.0.0.1:8282", nil)))
	def.Apply(nodes)

	// 随机选10次
	for i := 0; i < 10; i++ {
		selected, done, err := def.Select(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(selected.Scheme(), selected.Address())

		done(context.Background(), selector.DoneInfo{})
	}
}

func TestSelectNodeWithStaticNode(t *testing.T) {
	b := direct.Builder{}
	def := once.NewBuilder().Build()

	// 配置1个节点
	nodes := make([]selector.Node, 0, 1)
	nodes = append(nodes, b.Build(selector.NewNode("http", "127.0.0.1:8282", nil)))
	def.Apply(nodes)

	// 随机选10次
	for i := 0; i < 10; i++ {
		selected, done, err := def.Select(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(selected.Scheme(), selected.Address())

		done(context.Background(), selector.DoneInfo{})
	}
}

func TestSelectNodeWithP2C(t *testing.T) {
	b := direct.Builder{}
	def := p2c.NewBuilder().Build()
	// 配置10个节点
	nodes := make([]selector.Node, 0, 10)
	for i := 0; i < 10; i++ {
		nodes = append(nodes, b.Build(selector.NewNode("http", fmt.Sprintf("127.0.0.1:80%02d", i), nil)))
	}
	def.Apply(nodes)
	// 随机选10次
	for i := 0; i < 10; i++ {
		selected, done, err := def.Select(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(selected.Scheme(), selected.Address())

		done(context.Background(), selector.DoneInfo{})
	}
}
