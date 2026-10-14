package runtime

import (
	"context"
	"time"
)

// Wait returns when the delay expires or shutdown is requested.
func Wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
