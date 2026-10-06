package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/response"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	sessionstore "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/tests/httpcontract"
)

const contractID = "11111111-1111-4111-8111-111111111111"

func TestEveryPublishedOperationIsProxiedWithItsDeclaredShape(t *testing.T) {
	document := contract.Document("api")
	for path, item := range document.Paths.Map() {
		owner, _ := item.Extensions["x-owner"].(string)
		if owner == "" {
			continue
		}
		for method, operation := range item.Operations() {
			t.Run(operation.OperationID, func(t *testing.T) {
				urlPath := strings.ReplaceAll(path, "{id}", contractID)
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != strings.TrimPrefix(urlPath, "/api") || r.Method != method {
						t.Error("proxy route changed")
					}
					if r.Header.Get("Authorization") != "Bearer internal" || r.Header.Get("Cookie") != "" {
						t.Error("proxy credential isolation failed")
					}
					if method == "POST" && (r.Header.Get("Idempotency-Key") != contractID || r.Header.Get("If-Match") != "0" || r.Header.Get("X-Correlation-ID") != contractID) {
						t.Error("command metadata changed")
					}
					response.JSON(w, 200, operation.Responses.Value("200").Value.Content["application/json"].Example)
				}))
				defer upstream.Close()
				c := config()
				c.Backends = map[string]Backend{owner: {URL: upstream.URL, Key: "internal"}}
				var body []byte
				if operation.RequestBody != nil {
					body, _ = json.Marshal(operation.RequestBody.Value.Content["application/json"].Example)
				}
				r := httptest.NewRequest(method, urlPath, strings.NewReader(string(body)))
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Cookie", "cafe_session=valid")
				r.Header.Set("Origin", "http://cafe.local")
				for _, header := range []string{"Idempotency-Key", "X-Correlation-ID"} {
					r.Header.Set(header, contractID)
				}
				r.Header.Set("If-Match", "0")
				check.Request(t, document, r)
				w := httptest.NewRecorder()
				c.Handler().ServeHTTP(w, r)
				if w.Code != 200 {
					t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
				}
				check.Response(t, document, r, w.Result())
			})
		}
	}
}

func TestSessionAndTechnicalOperationsConformToOpenAPI(t *testing.T) {
	c := config()
	c.Diagnostics = func(context.Context) (any, error) { return sessionstore.Diagnostics{PendingRevocations: 1}, nil }
	for _, scenario := range []struct{ method, path, body string }{
		{"GET", "/healthz", ""}, {"GET", "/diagnostics", ""}, {"GET", "/auth/session", ""},
		{"POST", "/auth/login", `{"password":"code"}`}, {"POST", "/auth/logout", ""},
		{"POST", "/api/realtime/connect", `{}`}, {"POST", "/api/realtime/refresh", `{}`},
	} {
		t.Run(scenario.path, func(t *testing.T) {
			r := httptest.NewRequest(scenario.method, scenario.path, strings.NewReader(scenario.body))
			for key, value := range map[string]string{"Content-Type": "application/json", "Authorization": "Bearer cli", "Cookie": "cafe_session=valid", "Origin": "http://cafe.local", "X-Cafe-Realtime-Proxy": "proxy"} {
				r.Header.Set(key, value)
			}
			check.Request(t, contract.Document("api"), r)
			w := httptest.NewRecorder()
			c.Handler().ServeHTTP(w, r)
			if w.Code != 200 {
				t.Fatalf("status=%d", w.Code)
			}
			check.Response(t, contract.Document("api"), r, w.Result())
		})
	}
}

func TestAPIExposesOnlyOpenAPIOperations(t *testing.T) {
	called := false
	backend := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer backend.Close()
	c := config()
	c.Backends = map[string]Backend{"menu": {URL: backend.URL, Key: "internal"}}
	handler := c.Handler()
	for _, scenario := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/api/v1/menu/undocumented", 404}, {"POST", "/api/v1/menu/drinks", 405},
		{"DELETE", "/api/v1/menu/drinks/" + contractID, 405}, {"GET", "/api/v1/menu/drinks/" + contractID + "/extra", 404},
	} {
		r := httptest.NewRequest(scenario.method, scenario.path, nil)
		r.Header.Set("Authorization", "Bearer cli")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != scenario.status || called {
			t.Fatalf("undeclared route forwarded: %s %s status=%d", scenario.method, scenario.path, w.Code)
		}
	}
}
