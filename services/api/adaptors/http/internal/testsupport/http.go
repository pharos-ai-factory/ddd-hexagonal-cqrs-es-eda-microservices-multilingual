package testsupport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/tests/httpcontract"
)

type Sessions struct{}

func (Sessions) Create(context.Context) (string, error) { return "opaque", nil }
func (Sessions) Revoke(context.Context, string) error   { return nil }
func (Sessions) Authenticate(_ context.Context, token string) (a.Principal, bool, error) {
	return a.Principal{Subject: a.OperatorID, Role: "operator"}, token == "valid", nil
}
func Request(t *testing.T, handler http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	check.Response(t, contract.Document("api"), r, w.Result())
	return w
}

type UnavailableSessions struct{}

func (UnavailableSessions) Create(context.Context) (string, error) {
	return "", errors.New("session credential: private-secret")
}
func (UnavailableSessions) Authenticate(context.Context, string) (a.Principal, bool, error) {
	return a.Principal{}, false, errors.New("session credential: private-secret")
}
func (UnavailableSessions) Revoke(context.Context, string) error {
	return errors.New("session credential: private-secret")
}
