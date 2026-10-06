package session

import (
	"net/http"
	"testing"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func TestSessionFailuresRetainPublicErrorsAndCookies(t *testing.T) {
	c := config()
	c.Sessions = test.UnavailableSessions{}
	handler := newHandler(c)
	for _, scenario := range []struct{ method, path, body string }{
		{"POST", "/auth/login", `{"password":"code"}`},
		{"GET", "/auth/session", ""},
		{"POST", "/auth/logout", ""},
	} {
		t.Run(scenario.path, func(t *testing.T) {
			w := request(t, handler, scenario.method, scenario.path, scenario.body,
				map[string]string{"Cookie": "cafe_session=valid", "Origin": "http://cafe.local"})
			if w.Code != http.StatusServiceUnavailable || w.Body.String() != "{\"code\":\"session_unavailable\"}\n" {
				t.Fatalf("session failure status=%d body=%s", w.Code, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
				t.Fatal("failed session operation changed cookies or allowed response caching")
			}
		})
	}
}
