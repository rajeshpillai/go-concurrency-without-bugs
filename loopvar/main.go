// The loop variable: before Go 1.22, every goroutine saw the same i.
package main

import (
	"fmt"
	"slices"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var seen []int
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			seen = append(seen, i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	slices.Sort(seen)
	fmt.Println("the goroutines saw:", seen)
}
