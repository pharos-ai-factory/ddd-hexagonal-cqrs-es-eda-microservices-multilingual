package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	s "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

type orderWorld struct {
	orders s.CommandProbe[d.OrderState]
	menus  s.ProjectionProbe[model.MenuPublished]
}

func (w *orderWorld) add(id, edition, code string, quantity int) error {
	_, err := (app.AddLineCommandHandler{Orders: &w.orders, Menus: &w.menus}).Execute(context.Background(), s.Metadata(s.Order),
		app.AddLineCommand{LineID: id, EditionID: edition, OfferCode: code, Quantity: quantity})
	return err
}
func (w *orderWorld) quantity(quantity int) error {
	_, err := (app.ChangeQuantityCommandHandler{Orders: &w.orders}).Execute(context.Background(), s.Metadata(s.Order),
		app.ChangeQuantityCommand{LineID: s.FirstLine, Quantity: quantity})
	return err
}
func (w *orderWorld) place() error {
	_, err := (app.PlaceOrderCommandHandler{Orders: &w.orders}).Execute(context.Background(), s.Metadata(s.Order), struct{}{})
	return err
}
func TestOrderingFeatures(t *testing.T) {
	suite := godog.TestSuite{Name: "ordering",
		Options: s.FeatureOptions(t, "ordering", "../../../../../specifications/ordering", "@fast"),
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			w := &orderWorld{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*w = orderWorld{menus: s.ProjectionProbe[model.MenuPublished]{Values: map[string]model.MenuPublished{}}}
				return ctx, nil
			})
			sc.Step(`^Ordering knows a published edition offering "([^"]*)" for (\d+) minor units$`, func(name string, minor int64) {
				w.menus.Values[s.Edition] = model.MenuPublished{EditionID: s.Edition, Currency: "EUR", Offers: []model.Offer{
					{Code: "C1", DrinkID: s.Drink, DrinkRevision: 1, Name: name, Minor: minor, Currency: "EUR"}}}
			})
			sc.Step(`^a customer has a draft order for that edition$`, func() error {
				_, err := (app.CreateOrderCommandHandler{Orders: &w.orders, Menus: &w.menus}).Execute(context.Background(), s.Metadata(s.Order),
					app.CreateOrderCommand{CustomerID: s.Customer, EditionID: s.Edition})
				if err != nil {
					return err
				}
				return w.orders.Succeeded()
			})
			sc.Step(`^the order has a line containing (\d+) drinks?$`, func(n int) error {
				if err := w.add(s.FirstLine, s.Edition, "C1", n); err != nil {
					return err
				}
				return w.orders.Succeeded()
			})
			sc.Step(`^a second line contains (\d+) drinks$`, func(n int) error {
				if err := w.add(s.SecondLine, s.Edition, "C1", n); err != nil {
					return err
				}
				return w.orders.Succeeded()
			})
			sc.Step(`^the customer adds a second line containing (\d+) drinks?$`, func(n int) error { return w.add(s.SecondLine, s.Edition, "C1", n) })
			sc.Step(`^the customer changes the first line to (\d+) drinks$`, w.quantity)
			sc.Step(`^the customer places the order$`, w.place)
			sc.Step(`^the order is placed$`, func() error {
				if err := w.place(); err != nil {
					return err
				}
				return w.orders.Succeeded()
			})
			sc.Step(`^the order command is rejected with "([^"]*)"$`, w.orders.Rejected)
			sc.Step(`^the order command succeeds$`, w.orders.Succeeded)
			sc.Step(`^the order and its outgoing events are unchanged$`, w.orders.Unchanged)
			sc.Step(`^one public order publication contains (\d+) drinks at the frozen price$`, func(n int) error {
				if err := w.orders.Succeeded(); err != nil {
					return err
				}
				p := w.orders.Last.Publications
				if len(p) != 1 || p[0].Name != "ordering.order-placed" || p[0].Visibility != a.Public {
					return fmt.Errorf("wrong publications: %+v", p)
				}
				event, ok := p[0].Payload.(model.OrderPlaced)
				if !ok || event.OrderID != s.Order || event.CustomerID != s.Customer || event.EditionID != s.Edition || event.Currency != "EUR" || len(event.Lines) != 1 {
					return fmt.Errorf("wrong order: %+v", event)
				}
				line := event.Lines[0]
				if line.ID != s.FirstLine || line.Quantity != n || line.Minor != 300 || line.Name != "Coffee" || w.orders.Loaded.State.Status != "placed" {
					return fmt.Errorf("wrong frozen line: %+v", line)
				}
				return nil
			})
			sc.Step(`^the first line keeps its identity and contains (\d+) drinks$`, func(n int) error {
				if err := w.orders.Succeeded(); err != nil {
					return err
				}
				lines := w.orders.Loaded.State.Lines
				if len(lines) != 1 || lines[0].ID != s.FirstLine || lines[0].Quantity != n {
					return fmt.Errorf("wrong line: %+v", lines)
				}
				return nil
			})
			sc.Step(`^the customer attempts to "([^"]*)" the placed order$`, func(action string) error {
				switch action {
				case "change quantity":
					return w.quantity(2)
				case "add a line":
					return w.add(s.SecondLine, s.Edition, "C1", 1)
				}
				return fmt.Errorf("unknown action %q", action)
			})
			sc.Step(`^the customer selects an offer from a different edition$`, func() error { return w.add(s.FirstLine, s.Drink, "C1", 1) })
			sc.Step(`^the customer selects offer "([^"]*)"$`, func(code string) error { return w.add(s.FirstLine, s.Edition, code, 1) })
		}}
	if suite.Run() != 0 {
		t.Fatal("Ordering Gherkin scenarios failed")
	}
}
