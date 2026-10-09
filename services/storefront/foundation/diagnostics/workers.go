// Package diagnostics records technical worker failures without message bodies or credentials.
package diagnostics

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Failure records redacted operational failure evidence.
type Failure struct {
	At          time.Time `json:"at"`
	Class       string    `json:"class"`
	Event       string    `json:"event,omitempty"`
	Correlation string    `json:"correlation,omitempty"`
}

// Worker records process-local delivery counters and last failure.
type Worker struct {
	Owner               string  `json:"owner"`
	Name                string  `json:"name"`
	Failures            uint64  `json:"failures"`
	DeadLetterTransfers uint64  `json:"deadLetterTransfers"`
	LastFailure         Failure `json:"lastFailure"`
}

var workers = struct {
	sync.Mutex
	values map[string]Worker
}{values: map[string]Worker{}}

// Class deliberately excludes error messages: transport errors can contain URLs and secrets.
func Class(err error) string {
	if err == nil {
		return "lease_lost"
	}
	return fmt.Sprintf("%T", err)
}
func Record(owner, name, event, correlation string, err error, dead bool) {
	failure := Failure{time.Now().UTC(), Class(err), event, correlation}
	workers.Lock()
	key := owner + "/" + name
	worker := workers.values[key]
	worker.Owner, worker.Name = owner, name
	worker.Failures++
	if dead {
		worker.DeadLetterTransfers++
	}
	worker.LastFailure = failure
	workers.values[key] = worker
	workers.Unlock()
	slog.Warn("workflow failure", "owner", owner, "worker", name, "event", event,
		"correlation", correlation, "error_class", failure.Class, "dead_letter_transfer", dead)
}

// Snapshot contains counters since this process started, never durable queue depths.
func Snapshot() map[string]Worker {
	workers.Lock()
	defer workers.Unlock()
	result := make(map[string]Worker, len(workers.values))
	for key, worker := range workers.values {
		result[key] = worker
	}
	return result
}
