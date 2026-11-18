package retry

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// Inside a synctest bubble, time is fake: the 7 seconds of waiting take no real time.
func TestRetryWaits(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		calls := 0
		err := Do(context.Background(), 5, func() error {
			calls++
			if calls < 4 {
				return errors.New("server busy")
			}
			return nil
		})
		if err != nil || calls != 4 {
			t.Fatalf("err %v after %d calls", err, calls)
		}
		if waited := time.Since(start); waited != 7*time.Second {
			t.Fatalf("waited %s, want 7s", waited)
		}
		t.Logf("succeeded on try %d after waiting %s (fake time)", calls, time.Since(start))
	})
}
