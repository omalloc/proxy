package proxy

import (
	"net/http"
	"sync"

	"github.com/omalloc/proxy/selector"
)

var global = &proxyAppliance{}

type proxyAppliance struct {
	lock sync.Mutex

	Proxy
}

func init() {
	global.SetProxy(New())
}

func (a *proxyAppliance) SetProxy(in Proxy) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.Proxy = in
}

func SetProxy(in Proxy) {
	global.SetProxy(in)
}

func GetProxy() Proxy {
	return global
}

func Do(req *http.Request) (*http.Response, error) {
	return global.Do(req)
}

func Apply(nodes []selector.Node) {
	global.Apply(nodes)
}
