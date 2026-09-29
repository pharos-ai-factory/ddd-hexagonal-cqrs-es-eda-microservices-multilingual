package diagnostics

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestFailureBeforeClaimIsRetainedWithoutSecretMaterial(t *testing.T) {
	Record("test", "connection", "", "", errors.New("postgres://user:secret@host/private"), false)
	Record("test", "consumer", "event-id", "correlation-id", errors.New("body contains secret"), true)
	snapshot := Snapshot()
	if snapshot["test/connection"].Failures != 1 || snapshot["test/consumer"].DeadLetterTransfers != 1 {
		t.Fatal("failure evidence lost")
	}
	body, _ := json.Marshal(snapshot)
	if strings.Contains(string(body), "secret") || strings.Contains(string(body), "postgres://") {
		t.Fatal("sensitive error escaped")
	}
	delete(snapshot, "test/connection")
	if Snapshot()["test/connection"].Failures != 1 {
		t.Fatal("snapshot mutated process state")
	}
}
