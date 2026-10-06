package backend

import (
	"errors"
	"net/http"
	"testing"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

type unavailableTransport struct{}

func (unavailableTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("backend credential: private-secret")
}

func TestUnavailableBackendRetainsItsPublicError(t *testing.T) {
	c := config()
	proxy := newBackendProxy(Target{URL: "http://backend.local", Key: "private-secret"})
	proxy.Transport = unavailableTransport{}
	w := request(t, c.authoriseBackend(proxy), "GET", "/api/v1/menu/drinks", "", map[string]string{"Authorization": "Bearer cli"})
	if w.Code != http.StatusServiceUnavailable || w.Body.String() != "{\"code\":\"temporarily_unavailable\"}\n" {
		t.Fatalf("upstream failure status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSessionFailureDoesNotForward(t *testing.T) {
	c := config()
	c.Sessions = test.UnavailableSessions{}
	called := false
	handler := c.authoriseBackend(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	w := request(t, handler, "GET", "/api/v1/menu/drinks", "", map[string]string{"Cookie": "cafe_session=valid"})
	if w.Code != http.StatusServiceUnavailable || w.Body.String() != "{\"code\":\"session_unavailable\"}\n" || called {
		t.Fatalf("session failure status=%d body=%s forwarded=%t", w.Code, w.Body.String(), called)
	}
	if w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
		t.Fatal("failed authentication changed cookies or allowed response caching")
	}
}
