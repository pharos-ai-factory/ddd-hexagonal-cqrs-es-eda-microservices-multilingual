package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/httpcontract"
)

func TestProviderHTTPConformsToOpenAPI(t *testing.T) {
	root := t.TempDir()
	h := web.Auth("internal", handler(root))
	document := contract.Document("provider")
	id := "11111111-1111-4111-8111-111111111111"
	body := `{"id":"` + id + `","recipient":"` + id + `","subject":"Ready","body":"Coffee"}`
	send := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer internal")
		r.Header.Set("Idempotency-Key", id)
		if status == 200 {
			check.Request(t, document, r)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s status=%d body=%s", method, path, w.Code, w.Body.String())
		}
		check.Response(t, document, r, w.Result())
		return w
	}
	send("GET", "/messages", "", 200)
	send("POST", "/failures", `{"count":1,"lostResponses":0}`, 200)
	send("POST", "/messages", body, 503)
	first := send("POST", "/messages", body, 200)
	second := send("POST", "/messages", body, 200)
	if first.Body.String() != second.Body.String() {
		t.Fatal("provider retry changed acceptance")
	}
	send("POST", "/messages", strings.Replace(body, "Coffee", "Tea", 1), 409)
	send("POST", "/messages", `{`, 400)
	send("POST", "/messages", `{"id":"invalid"}`, 400)
	send("GET", "/messages", "", 200)
	if err := os.WriteFile(filepath.Join(root, id+".json"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	send("POST", "/messages", body, 500)
	r := httptest.NewRequest("GET", "/messages", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("anonymous provider access accepted")
	}
	check.Response(t, document, r, w.Result())
}
