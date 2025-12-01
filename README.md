# Proxy

A flexible HTTP reverse proxy library for Go with pluggable load balancing strategies. Inspired by [go-kratos](https://github.com/go-kratos/kratos).

## Features

- **Pluggable Load Balancing**: Supports various balancing strategies including Random, Weighted Round Robin (WRR), P2C (Power of Two Choices), and EWMA.
- **Connection Management**: Built-in connection pooling and timeout configurations.
- **Dynamic Node Management**: Easily update the list of backend nodes.
- **Context Support**: Pass peer information via context.

## Installation

```bash
go get -u github.com/omalloc/proxy
```

## Usage

### Basic Example

Here is a simple example of how to use the proxy client with a specific selector and initial nodes.

```go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/omalloc/proxy"
	"github.com/omalloc/proxy/selector"
	"github.com/omalloc/proxy/selector/random"
)

func main() {
	// Define backend nodes
	node1 := selector.NewNode("http", "127.0.0.1:8081", nil)
	node2 := selector.NewNode("http", "127.0.0.1:8082", nil)

	// Initialize proxy with Random selector
	proxyClient := proxy.New(
		proxy.WithSelector(random.NewBuilder()),
		proxy.WithInitialNodes([]selector.Node{node1, node2}),
	)

	// Create a request
	targetURL := "http://example.com/api/v1/resource"
	req, _ := http.NewRequest(http.MethodGet, targetURL, nil)

	// Execute request via proxy
	resp, err := proxyClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Dump response
	info, _ := httputil.DumpResponse(resp, false)
	log.Printf("Response Info: \n%v", string(info))
}
```

### Selectors

The library supports multiple load balancing algorithms located in the `selector` package and its subdirectories:

- **Random**: Randomly selects a node.
- **Round Robin (WRR)**: Weighted Round Robin.
- **P2C**: Power of Two Choices (Least Loaded).
- **EWMA**: Exponentially Weighted Moving Average.
- **ONCE**: Selects a single node for all requests.

To use a different selector, pass it to `proxy.WithSelector()`:

```go
import "github.com/omalloc/proxy/selector/wrr"

// ...

proxyClient := proxy.New(
    proxy.WithSelector(wrr.NewBuilder()),
    // ...
)
```

## License

MIT
