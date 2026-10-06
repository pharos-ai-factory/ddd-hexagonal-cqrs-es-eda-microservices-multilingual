package httpcontract

import (
	"net/http/httptest"
	"strings"
	"testing"

	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
)

func TestValidationRejectsWireDrift(t *testing.T) {
	document := contract.Document("storefront")
	path := "/v1/menu/drinks/11111111-1111-4111-8111-111111111111"
	for _, body := range []string{`{"name":7}`, `{"name":"Coffee","unexpected":true}`, `{}`} {
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
		r.Header.Set("If-Match", "0")
		if err := RequestError(document, r); err == nil {
			t.Fatalf("invalid request accepted: %s", body)
		}
	}
	for _, scenario := range []struct {
		status int
		body   string
	}{
		{200, `{"aggregateId":"11111111-1111-4111-8111-111111111111","version":"1","status":"draft"}`},
		{200, `{"aggregateId":"11111111-1111-4111-8111-111111111111","version":1}`},
		{422, `{"aggregateId":"11111111-1111-4111-8111-111111111111","version":1,"status":"draft"}`},
		{418, `{"code":"teapot"}`},
	} {
		r := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(scenario.status)
		_, _ = w.WriteString(scenario.body)
		if err := ResponseError(document, r, w.Result()); err == nil {
			t.Fatalf("invalid response accepted: %d %s", scenario.status, scenario.body)
		}
	}
}
