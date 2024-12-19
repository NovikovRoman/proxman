package proxman

import (
	"math/rand"
	"net/url"
	"sync"
	"time"
)

type List struct {
	sync.RWMutex
	items   map[string]*Proxy
	maxBusy int
}

func New(proxies ...*url.URL) (p *List) {
	p = &List{
		items: map[string]*Proxy{},
	}

	p.Add(proxies...)
	return
}

func (p *List) SetMaxBusy(num int) {
	p.maxBusy = num
}

func (p *List) MaxBusy() int {
	return p.maxBusy
}

func (p *List) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.items)
}

func (p *List) Num() int {
	return p.Len()
}

func (p *List) IsEmpty() bool {
	return p.Len() == 0
}

func (p *List) Add(u ...*url.URL) {
	p.Lock()
	defer p.Unlock()

	for _, uu := range u {
		if _, ok := p.items[uu.String()]; !ok {
			p.items[uu.String()] = &Proxy{
				url: uu,
			}
		}
	}
}

func (p *List) Contains(u *url.URL) (ok bool) {
	p.RLock()
	defer p.RUnlock()
	_, ok = p.items[u.String()]
	return
}

func (p *List) Remove(u ...*url.URL) {
	p.Lock()
	defer p.Unlock()
	for _, uu := range u {
		delete(p.items, uu.String())
	}
}

func (p *List) Acquire() (u *url.URL) {
	p.resetDeferRelease()

	// There is a limit on the number of proxies in use, or all proxies are occupied.
	if p.maxBusy > 0 && p.maxBusy <= p.NumBusy() || p.NumBusy()+p.NumBanned() >= p.Len() {
		return
	}

	// A Lock is used instead of an RLock to ensure that no proxy data is being added at the moment.
	p.Lock()
	defer p.Unlock()
	shuffleProxy := []*Proxy{}
	for _, item := range p.items {
		if item.isFree() {
			shuffleProxy = append(shuffleProxy, item)
		}
	}

	if len(shuffleProxy) > 0 {
		if len(shuffleProxy) > 1 { // random proxy
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			rnd.Shuffle(len(shuffleProxy), func(i, j int) {
				shuffleProxy[i], shuffleProxy[j] = shuffleProxy[j], shuffleProxy[i]
			})
		}

		shuffleProxy[0].setBusy(true)
		u = shuffleProxy[0].url
	}
	return
}

func (p *List) List() (proxies []*url.URL) {
	p.resetDeferRelease()

	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		proxies = append(proxies, item.url)
	}
	return
}

func (p *List) FreeList() (proxies []*url.URL) {
	p.resetDeferRelease()

	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isFree() {
			proxies = append(proxies, item.url)
		}
	}
	return
}

func (p *List) BusyList() (proxies []*url.URL) {
	p.resetDeferRelease()

	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isBusy() {
			proxies = append(proxies, item.url)
		}
	}
	return
}

func (p *List) BannedList() (banned []*url.URL) {
	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isBanned() {
			banned = append(banned, item.url)
		}
	}
	return
}

func (p *List) NumFree() (num int) {
	p.resetDeferRelease()

	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isFree() {
			num++
		}
	}
	return
}

func (p *List) NumBusy() (num int) {
	p.resetDeferRelease()

	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isBusy() {
			num++
		}
	}
	return
}

func (p *List) NumBanned() (num int) {
	p.RLock()
	defer p.RUnlock()
	for _, item := range p.items {
		if item.isBanned() {
			num++
		}
	}
	return
}

func (p *List) Banned(u *url.URL) (ban bool, reason string) {
	p.RLock()
	defer p.RUnlock()
	if _, exists := p.items[u.String()]; exists {
		ban = p.items[u.String()].ban
		reason = p.items[u.String()].reason
	}
	return
}

func (p *List) Free(u *url.URL) bool {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.items[u.String()]; ok {
		return !p.items[u.String()].busy
	}
	return false
}

func (p *List) Busy(u *url.URL) bool {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.items[u.String()]; ok {
		return p.items[u.String()].busy
	}
	return false
}

func (p *List) Ban(u *url.URL, reason string) (ok bool) {
	p.Lock()
	defer p.Unlock()
	if _, ok = p.items[u.String()]; ok {
		p.items[u.String()].setBan(true, reason)
	}
	return
}

func (p *List) Unban(u *url.URL) (ok bool) {
	p.Lock()
	defer p.Unlock()
	if _, ok = p.items[u.String()]; ok {
		p.items[u.String()].setBan(false, "")
	}
	return
}

func (p *List) Release(u *url.URL) (ok bool) {
	p.Lock()
	defer p.Unlock()
	if _, ok = p.items[u.String()]; ok {
		p.items[u.String()].release()
	}
	return
}

func (p *List) Replace(proxies ...*url.URL) {
	processed := map[string]*url.URL{}
	for _, proxy := range proxies {
		if p.Contains(proxy) {
			p.Unban(proxy)

		} else {
			p.Add(proxy)
		}

		processed[proxy.String()] = proxy
	}

	p.Lock()
	defer p.Unlock()
	for _, proxy := range p.items {
		if _, ok := processed[proxy.String()]; !ok {
			delete(p.items, proxy.String())
		}
	}
}

func (p *List) Clear() {
	p.Lock()
	defer p.Unlock()
	p.items = map[string]*Proxy{}
}

func (p *List) DeferRelease(u *url.URL, d time.Duration) (ok bool) {
	p.Lock()
	defer p.Unlock()
	if _, ok = p.items[u.String()]; !ok {
		return
	}
	p.items[u.String()].setDeferRelease(time.Now().Add(d))
	return
}

func (p *List) resetDeferRelease() {
	p.Lock()
	defer p.Unlock()
	for _, proxy := range p.items {
		if proxy.isDefer() && proxy.deferTime.Before(time.Now()) {
			proxy.release()
		}
	}
}
