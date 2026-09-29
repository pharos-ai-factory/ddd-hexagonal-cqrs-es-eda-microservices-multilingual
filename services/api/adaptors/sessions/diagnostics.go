package sessions

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type Failure struct {
	At        time.Time `json:"at"`
	Operation string    `json:"operation"`
	Class     string    `json:"class"`
}
type Diagnostics struct {
	PendingRevocations                int64    `json:"pendingRevocations"`
	ExpiredSessionsAwaitingRevocation int64    `json:"expiredSessionsAwaitingRevocation"`
	Failures                          uint64   `json:"processFailures"`
	LastFailure                       *Failure `json:"lastFailure,omitempty"`
}

func (s *Store) record(operation string, err error) {
	if err == nil {
		return
	}
	failure := Failure{time.Now().UTC(), operation, fmt.Sprintf("%T", err)}
	s.diagnosticsMu.Lock()
	s.failures++
	s.lastFailure = &failure
	s.diagnosticsMu.Unlock()
	slog.Warn("session worker failure", "worker", operation, "error_class", failure.Class)
}
func (s *Store) Diagnostics(ctx context.Context) (Diagnostics, error) {
	s.diagnosticsMu.Lock()
	result := Diagnostics{Failures: s.failures, LastFailure: s.lastFailure}
	s.diagnosticsMu.Unlock()
	pending := s.client.SCard(ctx, prefix+"disconnects")
	expired := s.client.ZCount(ctx, prefix+"expirations", "-inf", fmt.Sprint(time.Now().Unix()))
	if err := pending.Err(); err != nil {
		s.record("diagnostics", err)
		return result, err
	}
	if err := expired.Err(); err != nil {
		s.record("diagnostics", err)
		return result, err
	}
	result.PendingRevocations, result.ExpiredSessionsAwaitingRevocation = pending.Val(), expired.Val()
	return result, nil
}

// Keep the pending evidence when completion fails; a later worker will retry.
func (s *Store) complete(ctx context.Context, work, fence string) {
	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, prefix+"disconnects", work)
		pipe.Del(ctx, fence)
		return nil
	})
	s.record("complete", err)
}
