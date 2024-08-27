package proxy

import "sync"

var global = &proxyAppliance{}

type proxyAppliance struct {
	lock sync.Mutex

	Proxy
}

func init() {
	global.SetProxy(DefaultProxy)
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
