// Package watch counts goroutines while the crawler runs (sampled every millisecond).
package watch

import (
	"runtime"
	"sync/atomic"
	"time"
)

type Peak struct {
	max  atomic.Int64
	stop chan struct{}
}

func Goroutines() *Peak {
	p := &Peak{stop: make(chan struct{})}
	go func() {
		for {
			select {
			case <-p.stop:
				return
			case <-time.After(time.Millisecond):
				if n := int64(runtime.NumGoroutine()); n > p.max.Load() {
					p.max.Store(n)
				}
			}
		}
	}()
	return p
}

func (p *Peak) Stop() int64 { close(p.stop); return p.max.Load() }
