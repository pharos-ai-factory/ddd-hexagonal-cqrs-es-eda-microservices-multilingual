package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	sessionstore "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/sessions"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/tests/httpcontract"
)

const contractID = "11111111-1111-4111-8111-111111111111"

func TestEveryPublishedOperationTranslatesItsDeclaredShape(t *testing.T) {
	document := contract.Document("api")
	for path, item := range document.Paths.Map() {
		owner, _ := item.Extensions["x-owner"].(string)
		if owner == "" {
			continue
		}
		for method, operation := range item.Operations() {
			t.Run(operation.OperationID, func(t *testing.T) {
				urlPath := strings.ReplaceAll(path, "{id}", contractID)
				c := config()
				c.Owners = []string{owner}
				c.Requests = requestCaller(func(_ context.Context, request pb.Request) (pb.Reply, error) {
					name, _, err := rpc.Validate(request, owner)
					if err != nil || name != operation.OperationID {
						t.Fatal("OpenAPI operation disagrees with Protobuf request", err)
					}
					reply, err := pb.NewReply(owner, request.GetRequestId())
					if err != nil {
						t.Fatal(err)
					}
					example := operation.Responses.Value("200").Value.Content["application/json"].Example
					if method == "POST" {
						metadata := pb.Command(request).GetMetadata()
						if metadata.CommandId != contractID || metadata.CorrelationId != contractID || metadata.GetExpectedVersion() != 0 {
							t.Fatal("command metadata changed")
						}
						outcome := &pb.Outcome{}
						if err := rpc.FromObject(example, outcome); err != nil {
							t.Fatal(err)
						}
						if err := rpc.SetPayload(reply, "outcome", outcome); err != nil {
							t.Fatal(err)
						}
					} else {
						prefix := 3
						if strings.HasPrefix(name, "list") {
							prefix = 4
						}
						payloadName := strings.ToLower(name[prefix:prefix+1]) + name[prefix+1:]
						payload, err := rpc.NewPayload(reply, payloadName)
						if err != nil {
							t.Fatal(err)
						}
						value := example
						if prefix == 4 {
							value = map[string]any{"items": example, "paged": false}
						}
						if err := rpc.FromObject(value, payload); err != nil {
							t.Fatal(err)
						}
						if err := rpc.SetPayload(reply, payloadName, payload); err != nil {
							t.Fatal(err)
						}
					}
					return reply, nil
				})
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
	c := config()
	c.Owners = []string{"menu"}
	c.Requests = requestCaller(func(context.Context, pb.Request) (pb.Reply, error) { called = true; return nil, nil })
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

type requestCaller func(context.Context, pb.Request) (pb.Reply, error)

func (f requestCaller) Call(ctx context.Context, request pb.Request) (pb.Reply, error) {
	return f(ctx, request)
}
