package support

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/httpcontract"
)

func TestCompositionHTTPConformsToOpenAPI(t *testing.T) {
	service := StorefrontRuntime{Mux: contract.NewMux("storefront", nil), Databases: map[string]*postgres.ContextDatabase{}}
	service.mountHTTP()
	if !reflect.DeepEqual(service.Mux.Patterns(), []string{"GET /diagnostics", "GET /healthz"}) {
		t.Fatal("owner HTTP must contain only operational endpoints")
	}
	for _, path := range []string{"/healthz", "/diagnostics"} {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		service.Mux.Handler().ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s status=%d", path, w.Code)
		}
		check.Response(t, contract.Document("storefront"), r, w.Result())
	}
}

func TestBusinessRequestsHaveNoOwnerHTTPEntryPoint(t *testing.T) {
	service := StorefrontRuntime{Mux: contract.NewMux("storefront", nil), Databases: map[string]*postgres.ContextDatabase{}}
	service.mountHTTP()
	handler := web.Auth("internal", service.Mux.Handler())
	raw, err := os.ReadFile("../../../../contracts/services/api/http_api/api.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct{ Paths map[string]json.RawMessage }
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	count := 0
	for path := range document.Paths {
		if !strings.HasPrefix(path, "/api/v1/") {
			continue
		}
		count++
		path = strings.ReplaceAll(strings.TrimPrefix(path, "/api"), "{id}", "11111111-1111-4111-8111-111111111111")
		for _, method := range []string{"GET", "POST"} {
			r := httptest.NewRequest(method, path, nil)
			r.Header.Set("Authorization", "Bearer internal")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != 404 {
				t.Fatalf("%s %s status=%d", method, path, w.Code)
			}
		}
	}
	if count == 0 {
		t.Fatal("public business route catalogue is empty")
	}
	for _, item := range []struct {
		path   string
		status int
	}{{"/healthz", 200}, {"/diagnostics", 401}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", item.path, nil))
		if w.Code != item.status {
			t.Fatalf("anonymous %s status=%d", item.path, w.Code)
		}
	}
}
