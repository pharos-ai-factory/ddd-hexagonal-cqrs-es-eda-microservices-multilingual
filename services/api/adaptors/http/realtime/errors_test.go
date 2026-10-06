package realtime

import (
	"net/http"
	"strings"
	"testing"

	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
)

func TestSessionFailuresRetainCentrifugoEnvelopes(t *testing.T) {
	c := config()
	c.Sessions = test.UnavailableSessions{}
	handler := newHandler(c)
	for _, path := range []string{"/api/realtime/connect", "/api/realtime/refresh"} {
		w := request(t, handler, "POST", path, `{}`, map[string]string{
			"Cookie": "cafe_session=valid", "Origin": "http://cafe.local", "X-Cafe-Realtime-Proxy": "proxy"})
		if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "private-secret") {
			t.Fatalf("realtime authority failure changed protocol: %d %s", w.Code, w.Body.String())
		}
	}
}
