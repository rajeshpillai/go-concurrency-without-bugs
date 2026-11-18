// Package retry tries again after a failure, waiting longer each time.
package retry

import (
	"context"
	"time"
)

// Do calls f up to tries times. Between tries it waits 1s, then 2s, then 4s...
func Do(ctx context.Context, tries int, f func() error) error {
	wait := time.Second
	var err error
	for i := range tries {
		if err = f(); err == nil {
			return nil
		}
		if i == tries-1 {
			break
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
		wait *= 2
	}
	return err
}
