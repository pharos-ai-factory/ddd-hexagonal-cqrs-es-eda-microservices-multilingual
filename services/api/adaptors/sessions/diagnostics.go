package sessions

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Failure records redacted operational failure evidence.
type Failure struct {
	At        time.Time `json:"at"`
	Operation string    `json:"operation"`
	Class     string    `json:"class"`
}

// Diagnostics reports owned backlog and worker recovery evidence.
type Diagnostics struct {
	PendingRevocations                int64    `json:"pendingRevocations"`
	ExpiredSessionsAwaitingRevocation int64    `json:"expiredSessionsAwaitingRevocation"`
	Failures                          uint64   `json:"processFailures"`
	LastFailure                       *Failure `json:"lastFailure,omitempty"`
}

func (s *ValkeySessionStore) record(operation string, err error) {
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
func (s *ValkeySessionStore) Diagnostics(ctx context.Context) (Diagnostics, error) {
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
