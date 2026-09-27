package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

func TestCommandBoundaryRejectsAmbiguousBodiesAndMissingVersions(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	type input struct {
		Name string `json:"name"`
	}
	for _, test := range []struct {
		body, version string
		status        int
	}{
		{`null`, "0", 400}, {`[]`, "0", 400}, {`{}`, "", 428},
		{`{"name":"coffee","unknown":1}`, "0", 400},
		{`{"name":"coffee"} {}`, "0", 400},
		{`{"name":"coffee"}`, "0", 200},
	} {
		t.Run(test.body+test.version, func(t *testing.T) {
			called := false
			mux := http.NewServeMux()
			mux.HandleFunc("POST /commands/{id}", Command("test.Create", func(_ context.Context, m a.Metadata, value input) (a.Outcome, error) {
				called = true
				if m.AggregateID != id || m.ID != id || m.CorrelationID != id || m.ExpectedVersion == nil || *m.ExpectedVersion != 0 || value.Name != "coffee" {
					t.Fatalf("incorrect command metadata: %+v %+v", m, value)
				}
				return a.Outcome{AggregateID: id, Version: 1, Status: "created"}, nil
			}))
			request := httptest.NewRequest("POST", "/commands/"+id, strings.NewReader(test.body))
			request.Header.Set("Idempotency-Key", id)
			request.Header.Set("If-Match", test.version)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != test.status || called != (test.status == 200) {
				t.Fatalf("status=%d handler-called=%t", response.Code, called)
			}
		})
	}
}
