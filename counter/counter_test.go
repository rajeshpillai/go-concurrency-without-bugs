// Three ways to count from many goroutines at once: a mutex, an atomic, and a channel.
package counter

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	n := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			n++
			mu.Unlock()
		}
	})
}

func BenchmarkAtomic(b *testing.B) {
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Add(1)
		}
	})
}

// One goroutine owns the count; everyone else sends it a message.
func BenchmarkChannel(b *testing.B) {
	adds := make(chan struct{}, 128)
	done := make(chan int)
	go func() {
		n := 0
		for range adds {
			n++
		}
		done <- n
	}()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			adds <- struct{}{}
		}
	})
	close(adds)
	<-done
}
