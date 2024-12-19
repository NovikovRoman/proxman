package proxman

import (
	"net/url"
	"time"
)

type Proxy struct {
	url       *url.URL
	ban       bool
	reason    string
	busy      bool
	deferTime time.Time
}

func (p *Proxy) String() string {
	return p.url.String()
}

func (p *Proxy) release() {
	p.busy = false
	p.deferTime = time.Time{}
}

func (p *Proxy) setDeferRelease(t time.Time) {
	if p.isBusy() {
		p.deferTime = t
	}
}

func (p *Proxy) isFree() bool {
	return !p.ban && !p.busy
}

func (p *Proxy) setBusy(ok bool) {
	p.busy = ok
}

func (p *Proxy) isBusy() bool {
	return p.busy && !p.ban
}

func (p *Proxy) setBan(ok bool, reason string) {
	p.ban = ok
	p.busy = true
	p.reason = reason
}

func (p *Proxy) isBanned() bool {
	return p.ban
}

func (p *Proxy) isDefer() bool {
	return p.isBusy() && !p.deferTime.IsZero()
}
