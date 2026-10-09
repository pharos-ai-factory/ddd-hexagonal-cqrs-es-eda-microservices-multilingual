package httpcontract

import (
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	"net/http/httptest"
	"testing"
)

func TestValidationRejectsOperationalWireDrift(t *testing.T) {
	document := contract.Document("storefront")
	for _, scenario := range []struct {
		body  string
		valid bool
	}{
		{`{"status":"ok"}`, true}, {`{"status":7}`, false}, {`{}`, false},
		{`{"status":"ok","unexpected":true}`, false},
	} {
		r := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.WriteString(scenario.body)
		if err := ResponseError(document, r, w.Result()); (err == nil) != scenario.valid {
			t.Fatalf("response=%s valid=%t error=%v", scenario.body, scenario.valid, err)
		}
	}
}
