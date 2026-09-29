package http

import (
	"context"
	"encoding/base64"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPaginationParameters(t *testing.T) {
	if request, err := pageRequest(url.Values{}, "/drinks"); request != nil || err != nil {
		t.Fatal(request, err)
	}
	id := "00000000-0000-4000-8000-000000000001"
	cursor := base64.RawURLEncoding.EncodeToString([]byte("1|/drinks|" + id))
	request, err := pageRequest(url.Values{"limit": {"100"}, "cursor": {cursor}}, "/drinks")
	if err != nil || request.After != id || request.Limit != 100 {
		t.Fatal(request, err)
	}
	if _, err = pageRequest(url.Values{"limit": {"1"}, "cursor": {cursor}}, "/editions"); err == nil {
		t.Fatal("foreign cursor accepted")
	}
	for _, raw := range []string{"limit=0", "limit=101", "limit=-1", "limit=01", "limit=1.2", "limit=1&limit=2", "cursor=abc", "limit=1&cursor=", "limit=1&cursor=***", "limit=1&cursor=a&cursor=b"} {
		values, _ := url.ParseQuery(raw)
		if _, err := pageRequest(values, "/drinks"); err == nil {
			t.Fatal("accepted", raw)
		}
	}
}
func TestUnpaginatedQueriesRemainAvailable(t *testing.T) {
	handler := PagedList(func(context.Context) ([]a.Loaded[int], error) { return []a.Loaded[int]{{State: 1}}, nil },
		func(context.Context, a.PageRequest) (a.Page[int], error) {
			t.Fatal("unexpected pagination")
			return a.Page[int]{}, nil
		})
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest("GET", "/drinks", nil))
	if response.Code != 200 || response.Body.String()[0] != '[' {
		t.Fatal(response.Body.String())
	}
}
