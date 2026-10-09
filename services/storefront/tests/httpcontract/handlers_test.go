package httpcontract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	menuhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/http"
	menuapp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	orderhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/http"
	orderapp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	order "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/openapi"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/httpcontract"
	probe "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

const id = "11111111-1111-4111-8111-111111111111"
const other = "22222222-2222-4222-8222-222222222222"

type queries[S any] struct{ state S }

func (q queries[S]) Get(context.Context, string) (a.Loaded[S], error) {
	return a.Loaded[S]{Exists: true, Version: 1, State: q.state}, nil
}
func (q queries[S]) List(ctx context.Context) ([]a.Loaded[S], error) {
	item, _ := q.Get(ctx, id)
	return []a.Loaded[S]{item}, nil
}
func (q queries[S]) Page(ctx context.Context, request a.PageRequest) (a.Page[S], error) {
	items, _ := q.List(ctx)
	next := ""
	if request.After == "" {
		next = other
	}
	return a.Page[S]{Items: items, NextID: next}, nil
}

// These probes exercise actual application decisions and HTTP encoding.
// Receipt atomicity and durable delivery remain covered by the infrastructure lane.
func handlers(operation string) http.Handler {
	drink := menu.DrinkState{ID: id, Name: "Coffee", Revision: 1}
	offer := menu.OfferState{Code: "C1", DrinkID: other, DrinkRevision: 1, Name: "Coffee", Minor: 0, Currency: "EUR"}
	edition := menu.EditionState{ID: id, Currency: "EUR", Status: "draft", Offers: []menu.OfferState{offer}}
	state := order.OrderState{ID: id, CustomerID: other, EditionID: other, Currency: "EUR", Status: "draft", Lines: []order.LineState{{ID: other, Selection: order.Selection{OfferCode: "C1", Name: "Coffee", Minor: 0}, Quantity: 1}}}
	drinks := &probe.CommandProbe[menu.DrinkState]{Loaded: a.Loaded[menu.DrinkState]{Exists: true, Version: 1, State: drink}}
	editions := &probe.CommandProbe[menu.EditionState]{Loaded: a.Loaded[menu.EditionState]{Exists: true, Version: 1, State: edition}}
	orders := &probe.CommandProbe[order.OrderState]{Loaded: a.Loaded[order.OrderState]{Exists: true, Version: 1, State: state}}
	switch operation {
	case "createDrink":
		drinks.Loaded.Exists = false
	case "createEdition":
		editions.Loaded.Exists = false
	case "addOffer":
		editions.Loaded.State.Offers = []menu.OfferState{}
	case "createOrder":
		orders.Loaded.Exists = false
	case "addLine":
		orders.Loaded.State.Lines = []order.LineState{}
	}
	published := &probe.ProjectionProbe[menuapp.DrinkPublished]{Values: map[string]menuapp.DrinkPublished{other + "/1": {DrinkID: other, Name: "Coffee", Revision: 1}}}
	menus := &probe.ProjectionProbe[model.MenuPublished]{Values: map[string]model.MenuPublished{other: {EditionID: other, Currency: "EUR", Offers: []model.Offer{{Code: "C1", DrinkID: other, DrinkRevision: 1, Name: "Coffee", Minor: 0, Currency: "EUR"}}}}}
	mux := contract.NewMux("storefront", nil)
	menuhttp.Mount(mux, menuhttp.MenuHTTPHandlers{
		DrinkQueries:   menuapp.DrinkQueries{Read: queries[menu.DrinkState]{drink}},
		EditionQueries: menuapp.EditionQueries{Read: queries[menu.EditionState]{edition}},
		CreateDrink:    menucommands.CreateDrinkCommandHandler{Drinks: drinks}, ReviseDrink: menucommands.ReviseDrinkCommandHandler{Drinks: drinks}, PublishDrink: menucommands.PublishDrinkCommandHandler{Drinks: drinks},
		CreateEdition: menucommands.CreateEditionCommandHandler{Editions: editions}, AddOffer: menucommands.AddOfferCommandHandler{Editions: editions, Drinks: published}, ChangePrice: menucommands.ChangePriceCommandHandler{Editions: editions}, PublishEdition: menucommands.PublishEditionCommandHandler{Editions: editions},
	})
	orderhttp.Mount(mux, orderhttp.OrderingHTTPHandlers{
		OrderingQueries: orderapp.OrderingQueries{Read: queries[order.OrderState]{state}},
		CreateOrder:     orderingcommands.CreateOrderCommandHandler{Orders: orders, Menus: menus}, AddLine: orderingcommands.AddLineCommandHandler{Orders: orders, Menus: menus}, ChangeQuantity: orderingcommands.ChangeQuantityCommandHandler{Orders: orders}, PlaceOrder: orderingcommands.PlaceOrderCommandHandler{Orders: orders},
	})
	// Composition-owned technical handlers have their own conformance tests.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { web.JSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /diagnostics", func(w http.ResponseWriter, r *http.Request) {
		web.JSON(w, 503, map[string]string{"code": "diagnostics_unavailable"})
	})
	return web.Auth("internal", mux.Handler())
}

func TestEveryStorefrontOperationConformsToOpenAPI(t *testing.T) {
	document := contract.Document("storefront")
	for path, item := range document.Paths.Map() {
		if item.Extensions["x-owner"] == nil {
			continue
		}
		for method, operation := range item.Operations() {
			t.Run(operation.OperationID, func(t *testing.T) {
				var body []byte
				if operation.RequestBody != nil {
					body, _ = json.Marshal(operation.RequestBody.Value.Content["application/json"].Example)
				}
				paths := []string{strings.ReplaceAll(path, "{id}", id)}
				if method == "GET" && !strings.Contains(path, "{id}") {
					paths = append(paths, path+"?limit=1")
				}
				for index, urlPath := range paths {
					r := httptest.NewRequest(method, urlPath, strings.NewReader(string(body)))
					for key, value := range map[string]string{"Content-Type": "application/json", "Authorization": "Bearer internal", "Idempotency-Key": id, "If-Match": "1", "X-Correlation-ID": other} {
						r.Header.Set(key, value)
					}
					check.Request(t, document, r)
					w := httptest.NewRecorder()
					handler := handlers(operation.OperationID)
					handler.ServeHTTP(w, r)
					if w.Code != 200 {
						t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
					}
					check.Response(t, document, r, w.Result())
					if index == 1 {
						var page struct{ NextCursor string }
						if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.NextCursor == "" {
							t.Fatal("missing next cursor")
						}
						r = httptest.NewRequest("GET", path+"?limit=1&cursor="+page.NextCursor, nil)
						r.Header.Set("Authorization", "Bearer internal")
						check.Request(t, document, r)
						w = httptest.NewRecorder()
						handler.ServeHTTP(w, r)
						check.Response(t, document, r, w.Result())
						if !strings.Contains(w.Body.String(), `"nextCursor":null`) {
							t.Fatal("last cursor must be null")
						}
					}
				}
			})
		}
	}
}
