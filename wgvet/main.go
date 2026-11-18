// A WaitGroup bug that go vet catches (Go 1.25+): Add inside the goroutine it counts.
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := range 3 {
		go func() {
			wg.Add(1) // too late: Wait may already have returned
			defer wg.Done()
			fmt.Println("page", i)
		}()
	}
	wg.Wait()
}
