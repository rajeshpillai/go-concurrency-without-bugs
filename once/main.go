// sync.OnceValue: run an expensive setup exactly once, however many goroutines ask for it.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var loads atomic.Int64

// config loads the settings once; every later call returns the same value.
var config = sync.OnceValue(func() map[string]int {
	loads.Add(1)
	time.Sleep(100 * time.Millisecond) // pretend this reads a file
	return map[string]int{"workers": 8}
})

func main() {
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { _ = config()["workers"] })
	}
	wg.Wait()
	fmt.Printf("50 goroutines asked for the config. It was loaded %d time.\n", loads.Load())
}
