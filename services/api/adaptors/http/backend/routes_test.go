package backend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func TestAPIAuthenticatesAndReplacesBrowserCredentials(t *testing.T) {
	var forwarded http.Header
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = r.Header.Clone()
		if r.URL.Path != "/v1/menu/drinks/11111111-1111-4111-8111-111111111111" {
			t.Errorf("incorrect internal path: %s", r.URL.Path)
		}
		response.JSON(w, 200, map[string]any{"aggregateId": "11111111-1111-4111-8111-111111111111", "version": 1, "status": "draft"})
	}))
	defer backend.Close()
	c := config()
	c.Backends = map[string]Target{"menu": {URL: backend.URL, Key: "internal"}}
	handler := c.authoriseBackend(newBackendProxy(c.Backends["menu"]))
	if w := request(t, handler, "GET", "/api/v1/menu/drinks", "", nil); w.Code != 401 {
		t.Fatal("anonymous API access accepted")
	}
	if w := request(t, handler, "POST", "/api/v1/menu/drinks/11111111-1111-4111-8111-111111111111", `{}`, map[string]string{
		"Cookie": "cafe_session=valid", "Origin": "http://attacker.local"}); w.Code != 403 {
		t.Fatal("foreign origin mutation accepted")
	}
	w := request(t, handler, "POST", "/api/v1/menu/drinks/11111111-1111-4111-8111-111111111111", `{}`, map[string]string{
		"Cookie": "cafe_session=valid", "Origin": "http://cafe.local",
		"Authorization": "Bearer attacker", "Idempotency-Key": "command", "If-Match": "2"})
	if w.Code != 200 || forwarded.Get("Authorization") != "Bearer internal" ||
		forwarded.Get("Cookie") != "" || forwarded.Get("Idempotency-Key") != "command" {
		t.Fatal("credential isolation or command identity forwarding failed")
	}
}

func config() Config {
	return Config{Sessions: test.Sessions{}, CLIKey: "cli", Origins: []string{"http://cafe.local"}}
}

var request = test.Request
