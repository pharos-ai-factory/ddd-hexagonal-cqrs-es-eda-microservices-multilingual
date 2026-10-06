package support

import (
	"net/http/httptest"
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/httpcontract"
)

func TestCompositionHTTPConformsToOpenAPI(t *testing.T) {
	service := Service{Mux: contract.NewMux("storefront", map[string]bool{}), Databases: map[string]*postgres.Database{}}
	service.mountHTTP()
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
