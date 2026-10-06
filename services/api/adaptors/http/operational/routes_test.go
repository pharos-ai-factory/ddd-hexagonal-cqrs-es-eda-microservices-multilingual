package operational

import (
	"context"
	"errors"
	"strings"
	"testing"

	sessionstore "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
)

func TestDiagnosticsNeedsCLIKeyAndDoesNotChangeLiveness(t *testing.T) {
	c := config()
	called := false
	c.Diagnostics = func(context.Context) (any, error) {
		called = true
		return nil, errors.New("redis://user:secret@private")
	}
	handler := newHandler(c)
	if w := request(t, handler, "GET", "/diagnostics", "", nil); w.Code != 401 || called {
		t.Fatal("diagnostics exposed to unauthenticated caller")
	}
	if w := request(t, handler, "GET", "/diagnostics", "", map[string]string{"Authorization": "Bearer cli"}); w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal("unavailable diagnostics was not redacted")
	}
	if w := request(t, handler, "GET", "/healthz", "", nil); w.Code != 200 {
		t.Fatal("dependency failure incorrectly killed liveness")
	}
	c.Diagnostics = func(context.Context) (any, error) { return sessionstore.Diagnostics{PendingRevocations: 3}, nil }
	if w := request(t, newHandler(c), "GET", "/diagnostics", "", map[string]string{"Authorization": "Bearer cli"}); w.Code != 200 || !strings.Contains(w.Body.String(), `"pendingRevocations":3`) {
		t.Fatal("backlog missing")
	}
}
