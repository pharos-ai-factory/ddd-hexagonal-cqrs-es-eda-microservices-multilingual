package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

type sessions struct{}

func (sessions) Create(context.Context) (string, error) { return "opaque", nil }
func (sessions) Revoke(context.Context, string) error   { return nil }
func (sessions) Authenticate(_ context.Context, token string) (a.Principal, bool, error) {
	return a.Principal{Subject: a.OperatorID, Role: "operator"}, token == "valid", nil
}
func config() Config {
	return Config{Sessions: sessions{}, OperatorPassword: "code", CLIKey: "cli",
		ProxySecret: "proxy", Origins: []string{"http://cafe.local"}}
}
func request(handler http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}
func TestProxyChoosesChannelsAndRejectsUntrustedCallers(t *testing.T) {
	handler := config().Handler()
	if w := request(handler, "POST", "/api/realtime/connect", `{}`, nil); w.Code != 403 {
		t.Fatalf("untrusted proxy: %d", w.Code)
	}
	for _, cookie := range []string{"", "cafe_session=invalid"} {
		w := request(handler, "POST", "/api/realtime/connect", `{}`, map[string]string{
			"X-Cafe-Realtime-Proxy": "proxy", "Origin": "http://cafe.local", "Cookie": cookie})
		if !strings.Contains(w.Body.String(), `"disconnect"`) {
			t.Fatal("unauthenticated subscription accepted")
		}
	}
	w := request(handler, "POST", "/api/realtime/connect", `{"channels":["secret:foreign"]}`,
		map[string]string{"X-Cafe-Realtime-Proxy": "proxy", "Origin": "http://cafe.local", "Cookie": "cafe_session=valid"})
	var result struct {
		Result struct {
			User string
			Subs map[string]any
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Result.User != a.OperatorID || len(result.Result.Subs) != 6 {
		t.Fatal("missing server authority")
	}
	if _, ok := result.Result.Subs["secret:foreign"]; ok {
		t.Fatal("browser-selected subscription accepted")
	}
}
func TestAPIAuthenticatesAndReplacesBrowserCredentials(t *testing.T) {
	var forwarded http.Header
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = r.Header.Clone()
		if r.URL.Path != "/v1/menu/drinks" {
			t.Errorf("incorrect internal path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()
	c := config()
	c.Backends = map[string]Backend{"menu": {URL: backend.URL, Key: "internal"}}
	handler := c.Handler()
	if w := request(handler, "GET", "/api/v1/menu/drinks", "", nil); w.Code != 401 {
		t.Fatal("anonymous API access accepted")
	}
	if w := request(handler, "POST", "/api/v1/menu/drinks", `{}`, map[string]string{
		"Cookie": "cafe_session=valid", "Origin": "http://attacker.local"}); w.Code != 403 {
		t.Fatal("foreign origin mutation accepted")
	}
	w := request(handler, "POST", "/api/v1/menu/drinks", `{}`, map[string]string{
		"Cookie": "cafe_session=valid", "Origin": "http://cafe.local",
		"Authorization": "Bearer attacker", "Idempotency-Key": "command", "If-Match": "2"})
	if w.Code != 200 || forwarded.Get("Authorization") != "Bearer internal" ||
		forwarded.Get("Cookie") != "" || forwarded.Get("Idempotency-Key") != "command" {
		t.Fatal("credential isolation or command identity forwarding failed")
	}
}
func TestLoginRequiresOriginAndSetsPrivateCookie(t *testing.T) {
	handler := config().Handler()
	if w := request(handler, "POST", "/auth/login", `{"password":"code"}`, nil); w.Code != 403 {
		t.Fatal("cross-origin login accepted")
	}
	w := request(handler, "POST", "/auth/login", `{"password":"code"}`, map[string]string{"Origin": "http://cafe.local"})
	cookies := w.Result().Cookies()
	if w.Code != 200 || len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("session cookie policy missing")
	}
}
