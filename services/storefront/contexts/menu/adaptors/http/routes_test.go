package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	menuhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/http"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	s "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

func TestPriceCommandRequiresAnExplicitNonNullAmount(t *testing.T) {
	for _, body := range []string{`{"code":"C1"}`, `{"code":"C1","minor":null}`, `{"code":"C1","minor":0}`} {
		t.Run(body, func(t *testing.T) {
			editions := s.CommandProbe[d.EditionState]{Loaded: a.Loaded[d.EditionState]{
				Exists: true, Version: 1, State: d.EditionState{ID: s.Edition, Currency: "EUR", Status: "draft",
					Offers: []d.OfferState{{Code: "C1", DrinkID: s.Drink, DrinkRevision: 1,
						Name: "Coffee", Minor: 350, Currency: "EUR"}}},
			}}
			mux := http.NewServeMux()
			menuhttp.Mount(mux, menuhttp.MenuHTTPHandlers{ChangePrice: menucommands.ChangePriceCommandHandler{Editions: &editions}})
			request := httptest.NewRequest("POST", "/v1/menu/editions/"+s.Edition+"/prices", strings.NewReader(body))
			request.Header.Set("Idempotency-Key", s.Customer)
			request.Header.Set("If-Match", "1")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			wantStatus, wantPrice, wantVersion := http.StatusBadRequest, int64(350), uint64(1)
			if body == `{"code":"C1","minor":0}` {
				wantStatus, wantPrice, wantVersion = http.StatusOK, 0, 2
			}
			if response.Code != wantStatus {
				t.Errorf("status=%d, want %d", response.Code, wantStatus)
			}
			if price := editions.Loaded.State.Offers[0].Minor; price != wantPrice || editions.Loaded.Version != wantVersion {
				t.Errorf("price=%d version=%d; incomplete prices must preserve the existing offer", price, editions.Loaded.Version)
			}
		})
	}
}

func TestAddOfferRequiresAnExplicitNonNullAmount(t *testing.T) {
	for _, amount := range []string{"", `,"minor":null`, `,"minor":0`} {
		t.Run(amount, func(t *testing.T) {
			editions := s.CommandProbe[d.EditionState]{Loaded: a.Loaded[d.EditionState]{
				Exists: true, Version: 1, State: d.EditionState{ID: s.Edition, Currency: "EUR", Status: "draft"},
			}}
			drinks := s.ProjectionProbe[app.DrinkPublished]{Values: map[string]app.DrinkPublished{
				s.Drink + "/1": {DrinkID: s.Drink, Name: "Coffee", Revision: 1},
			}}
			mux := http.NewServeMux()
			menuhttp.Mount(mux, menuhttp.MenuHTTPHandlers{AddOffer: menucommands.AddOfferCommandHandler{Editions: &editions, Drinks: &drinks}})
			body := `{"code":"C1","drinkId":"` + s.Drink + `","drinkRevision":1` + amount + `}`
			request := httptest.NewRequest("POST", "/v1/menu/editions/"+s.Edition+"/offers", strings.NewReader(body))
			request.Header.Set("Idempotency-Key", s.Customer)
			request.Header.Set("If-Match", "1")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if amount == `,"minor":0` {
				if response.Code != 200 || len(editions.Loaded.State.Offers) != 1 ||
					editions.Loaded.State.Offers[0].Minor != 0 || editions.Loaded.Version != 2 {
					t.Fatal("an explicitly free offer was not added")
				}
			} else if response.Code != 400 || len(editions.Loaded.State.Offers) != 0 || editions.Loaded.Version != 1 {
				t.Fatal("an omitted or null price reached the offer decision")
			}
		})
	}
}
