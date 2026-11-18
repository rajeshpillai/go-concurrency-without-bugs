package watch

import (
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

// Stuck waits d; if the program is still running then, it prints what each goroutine is
// waiting for (counted by state, from the goroutine profile) and exits.
func Stuck(d time.Duration, done <-chan struct{}) {
	select {
	case <-done:
		return
	case <-time.After(d):
	}
	var b strings.Builder
	pprof.Lookup("goroutine").WriteTo(&b, 2)
	states := map[string]int{}
	for _, l := range strings.Split(b.String(), "\n") {
		if strings.HasPrefix(l, "goroutine ") {
			s := l[strings.Index(l, "[")+1 : strings.Index(l, "]")]
			if i := strings.Index(s, ","); i > 0 {
				s = s[:i]
			}
			states[s]++
		}
	}
	fmt.Printf("still running after %s. goroutines by what they wait for:\n", d)
	for s, n := range states {
		fmt.Printf("  %3d  %s\n", n, s)
	}
	os.Exit(1)
}
