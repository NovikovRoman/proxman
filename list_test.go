package proxman

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	pl := New()
	assert.True(t, pl.IsEmpty())
	assert.Nil(t, pl.Acquire())

	proxies := []*url.URL{p1, p2, p3, p4}
	pl = New(proxies...)
	assert.False(t, pl.IsEmpty())
	assert.Equal(t, len(proxies), pl.Len())

	assert.True(t, pl.Contains(p3))
	assert.False(t, pl.Contains(p5))
	pl.Add(p5)
	assert.True(t, pl.Contains(p5))
	assert.Equal(t, pl.Len(), len(proxies)+1)
	pl.Add(p5)
	assert.Equal(t, pl.Len(), len(proxies)+1)

	pl.Remove(p1)
	assert.Equal(t, pl.Len(), len(proxies))
	assert.Equal(t, pl.Num(), len(proxies))
	assert.False(t, pl.Contains(p1))

	pl.Clear()
	assert.True(t, pl.IsEmpty())
}

func TestFreeBusyBanned(t *testing.T) {
	proxies := []*url.URL{p1, p2, p3, p4}
	pl := New(proxies...)

	pBusy := pl.Acquire()
	freeList := pl.FreeList()

	assert.Equal(t, pl.NumFree(), len(proxies)-1)
	assert.Equal(t, len(freeList), len(proxies)-1)
	assert.Equal(t, pl.NumBusy(), 1)
	assert.Equal(t, pBusy, pl.BusyList()[0])
	assert.Equal(t, pl.NumBanned(), 0)
	assert.Equal(t, len(pl.BannedList()), 0)
	assert.Equal(t, len(pl.List()), len(proxies))
	assert.False(t, pl.Free(pBusy))
	assert.True(t, pl.Free(freeList[0]))
	assert.False(t, pl.Free(p5))
	assert.False(t, pl.Contains(p5))

	assert.True(t, pl.Busy(pBusy))
	assert.False(t, pl.Busy(p5))

	pBan := freeList[0]
	reason := "test reason"
	assert.False(t, pl.Ban(p5, reason))
	assert.True(t, pl.Ban(pBan, reason))

	ok, reasonBan := pl.Banned(pBan)
	assert.True(t, ok)
	assert.Equal(t, reason, reasonBan)
	assert.Equal(t, pl.NumFree(), len(proxies)-2)
	assert.Equal(t, len(pl.FreeList()), len(proxies)-2)
	assert.Equal(t, pl.NumBusy(), 1)
	assert.Equal(t, pBusy, pl.BusyList()[0])
	assert.Equal(t, pl.NumBanned(), 1)
	assert.Equal(t, len(pl.BannedList()), 1)
	assert.Equal(t, len(pl.List()), len(proxies))
	assert.Equal(t, pBan, pl.BannedList()[0])

	ok = pl.Unban(p5)
	assert.False(t, ok)
	ok = pl.Unban(pBan)
	assert.True(t, ok)

	ok = pl.Release(pBusy)
	assert.True(t, ok)
	ok = pl.Release(p5)
	assert.False(t, ok)
}

func TestRelease(t *testing.T) {
	proxies := []*url.URL{p1, p2, p3, p4}
	pl := New(proxies...)
	pBusy1 := pl.Acquire()
	assert.Equal(t, 3, pl.NumFree())
	pBusy2 := pl.Acquire()
	assert.Equal(t, 2, pl.NumFree())

	ok := pl.Release(pBusy1)
	assert.True(t, ok)
	ok = pl.Release(p5)
	assert.False(t, ok)
	assert.Equal(t, pl.NumFree(), 3)

	ok = pl.DeferRelease(pBusy2, time.Second/4)
	assert.True(t, ok)
	ok = pl.DeferRelease(p5, time.Second/4)
	assert.False(t, ok)
	assert.Equal(t, pl.NumFree(), 3)
	time.Sleep(time.Second)
	assert.Equal(t, pl.NumFree(), 4)
}

func TestMaxBusy(t *testing.T) {
	proxies := []*url.URL{p1, p2, p3}
	pl := New(proxies...)
	assert.Equal(t, pl.MaxBusy(), 0)
	p1 := pl.Acquire()
	p2 := pl.Acquire()
	p3 := pl.Acquire()
	assert.Equal(t, pl.NumBusy(), 3)
	assert.Nil(t, pl.Acquire())
	pl.Release(p1)
	pl.Release(p2)
	pl.Release(p3)

	pl.SetMaxBusy(2)
	assert.Equal(t, pl.MaxBusy(), 2)
	p1 = pl.Acquire()
	p2 = pl.Acquire()
	assert.Equal(t, pl.NumBusy(), 2)
	assert.Nil(t, pl.Acquire())
	pl.Release(p1)
	pl.Release(p2)

	pl.SetMaxBusy(1)
	assert.Equal(t, pl.MaxBusy(), 1)
	pl.Acquire()
	assert.Equal(t, pl.NumBusy(), 1)
	assert.Nil(t, pl.Acquire())
}

func TestReplace(t *testing.T) {
	proxies := []*url.URL{p1, p2}
	pl := New(proxies...)
	assert.Equal(t, pl.Len(), 2)
	pl.Replace(p1, p3)
	assert.Equal(t, pl.Len(), 2)
	assert.True(t, pl.Contains(p1))
	assert.False(t, pl.Contains(p2))
	assert.True(t, pl.Contains(p3))

	pl.Ban(p3, "test")
	pl.Replace(p3, p4)
	assert.Equal(t, pl.Len(), 2)
	assert.False(t, pl.Contains(p1))
	assert.True(t, pl.Contains(p3))
	assert.True(t, pl.Contains(p4))
	yes, _ := pl.Banned(p3)
	assert.False(t, yes)
}
