package backend

import (
	"context"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	"net/http"
	"testing"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func TestUnavailableBackendRetainsItsPublicError(t *testing.T) {
	c := config()
	c.Owners = []string{"menu"}
	c.Requests = caller(func(context.Context, pb.Request) (pb.Reply, error) { return nil, unavailable })
	w := request(t, backendHandler(c), "GET", "/api/v1/menu/drinks", "", map[string]string{"Authorization": "Bearer cli"})
	if w.Code != 503 || w.Body.String() != "{\"code\":\"temporarily_unavailable\"}\n" {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
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
