package http

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDiagnosticsNeedsCLIKeyAndDoesNotChangeLiveness(t *testing.T) {
	c := config()
	called := false
	c.Diagnostics = func(context.Context) (any, error) {
		called = true
		return nil, errors.New("redis://user:secret@private")
	}
	handler := c.Handler()
	if w := request(handler, "GET", "/diagnostics", "", nil); w.Code != 401 || called {
		t.Fatal("diagnostics exposed to unauthenticated caller")
	}
	if w := request(handler, "GET", "/diagnostics", "", map[string]string{"Authorization": "Bearer cli"}); w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal("unavailable diagnostics was not redacted")
	}
	if w := request(handler, "GET", "/healthz", "", nil); w.Code != 200 {
		t.Fatal("dependency failure incorrectly killed liveness")
	}
	c.Diagnostics = func(context.Context) (any, error) { return map[string]int{"pendingRevocations": 3}, nil }
	if w := request(c.Handler(), "GET", "/diagnostics", "", map[string]string{"Authorization": "Bearer cli"}); w.Code != 200 || !strings.Contains(w.Body.String(), `"pendingRevocations":3`) {
		t.Fatal("backlog missing")
	}
}
