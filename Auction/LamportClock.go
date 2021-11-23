package Auction

import "sync"

type LamportClock struct {
	clock uint32
	lock  sync.Mutex
}

func (c *LamportClock) Tick() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.clock += 1
}
func (c *LamportClock) SyncClocks(otherClock uint32) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.clock = Max(c.clock, otherClock) + 1
}
func (c *LamportClock) GetTime() uint32 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.clock
}

func Max(a uint32, b uint32) uint32 {
	if b > a {
		return b
	}
	return a
}
