package commands_test

import (
	"context"
	"fmt"
	menupub "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/publications"
	menuports "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	"testing"

	"github.com/cucumber/godog"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	s "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

type menuWorld struct {
	drinks    s.CommandProbe[d.DrinkState]
	editions  s.CommandProbe[d.EditionState]
	directory s.ProjectionProbe[app.DrinkPublished]
}

func (w *menuWorld) publishDrink() error {
	_, err := (execution.Bind(s.ForAggregate(&w.drinks, d.RestoreDrink, (*d.Drink).Snapshot, menupub.Drink), func(repository menuports.DrinkWriteRepository) menucommands.PublishDrinkCommandHandler {
		return menucommands.PublishDrinkCommandHandler{Repository: repository}
	})).Execute(context.Background(), s.Metadata(s.Drink), struct{}{})
	if err != nil {
		return err
	}
	if err = w.drinks.Succeeded(); err != nil {
		return err
	}
	state := w.drinks.Loaded.State
	w.directory.Values[fmt.Sprintf("%s/%d", s.Drink, state.Revision)] = app.DrinkPublished{DrinkID: s.Drink, Name: state.Name, Revision: state.Revision}
	return nil
}
func (w *menuWorld) add(code string, minor int64) error {
	_, err := (execution.Bind(s.ForAggregate(&w.editions, d.RestoreEdition, (*d.MenuEdition).Snapshot, menupub.Edition), func(repository menuports.EditionWriteRepository) menucommands.AddOfferCommandHandler {
		return menucommands.AddOfferCommandHandler{Repository: repository, Drinks: &w.directory}
	})).Execute(context.Background(), s.Metadata(s.Edition),
		menucommands.AddOfferCommand{Code: code, DrinkID: s.Drink, DrinkRevision: 1, Minor: minor})
	return err
}
func (w *menuWorld) publish() error {
	_, err := (execution.Bind(s.ForAggregate(&w.editions, d.RestoreEdition, (*d.MenuEdition).Snapshot, menupub.Edition), func(repository menuports.EditionWriteRepository) menucommands.PublishEditionCommandHandler {
		return menucommands.PublishEditionCommandHandler{Repository: repository}
	})).Execute(context.Background(), s.Metadata(s.Edition), struct{}{})
	return err
}
func (w *menuWorld) price(code string, minor int64) error {
	_, err := (execution.Bind(s.ForAggregate(&w.editions, d.RestoreEdition, (*d.MenuEdition).Snapshot, menupub.Edition), func(repository menuports.EditionWriteRepository) menucommands.ChangePriceCommandHandler {
		return menucommands.ChangePriceCommandHandler{Repository: repository}
	})).Execute(context.Background(), s.Metadata(s.Edition), menucommands.ChangePriceCommand{Code: code, Minor: minor})
	return err
}
func (w *menuWorld) offer() (d.OfferState, error) {
	if len(w.editions.Loaded.State.Offers) != 1 {
		return d.OfferState{}, fmt.Errorf("expected one offer: %+v", w.editions.Loaded)
	}
	return w.editions.Loaded.State.Offers[0], nil
}

func TestMenuFeatures(t *testing.T) {
	suite := godog.TestSuite{Name: "menu",
		Options: s.FeatureOptions(t, "menu", "../../../../../../specifications/menu", "@fast"),
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			w := &menuWorld{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				*w = menuWorld{directory: s.ProjectionProbe[app.DrinkPublished]{Values: map[string]app.DrinkPublished{}}}
				return ctx, nil
			})
			sc.Step(`^a published drink named "([^"]*)"$`, func(name string) error {
				_, err := (execution.Bind(s.ForAggregate(&w.drinks, d.RestoreDrink, (*d.Drink).Snapshot, menupub.Drink), func(repository menuports.DrinkWriteRepository) menucommands.CreateDrinkCommandHandler {
					return menucommands.CreateDrinkCommandHandler{Repository: repository}
				})).Execute(context.Background(), s.Metadata(s.Drink), menucommands.CreateDrinkCommand{Name: name})
				if err != nil {
					return err
				}
				if err = w.drinks.Succeeded(); err != nil {
					return err
				}
				return w.publishDrink()
			})
			sc.Step(`^a draft menu edition in "([^"]*)"$`, func(currency string) error {
				_, err := (execution.Bind(s.ForAggregate(&w.editions, d.RestoreEdition, (*d.MenuEdition).Snapshot, menupub.Edition), func(repository menuports.EditionWriteRepository) menucommands.CreateEditionCommandHandler {
					return menucommands.CreateEditionCommandHandler{Repository: repository}
				})).Execute(context.Background(), s.Metadata(s.Edition), menucommands.CreateEditionCommand{Currency: currency})
				if err != nil {
					return err
				}
				return w.editions.Succeeded()
			})
			sc.Step(`^the edition offers "([^"]*)" for (\d+) minor units$`, func(code string, minor int64) error {
				if err := w.add(code, minor); err != nil {
					return err
				}
				return w.editions.Succeeded()
			})
			sc.Step(`^the operator adds offer "([^"]*)" for (\d+) minor units$`, w.add)
			sc.Step(`^the operator publishes the menu edition$`, w.publish)
			sc.Step(`^the menu edition is published$`, func() error {
				if err := w.publish(); err != nil {
					return err
				}
				return w.editions.Succeeded()
			})
			sc.Step(`^the menu command is rejected with "([^"]*)"$`, w.editions.Rejected)
			sc.Step(`^the menu command succeeds$`, w.editions.Succeeded)
			sc.Step(`^the edition and its outgoing events are unchanged$`, w.editions.Unchanged)
			sc.Step(`^the operator changes offer "([^"]*)" to (\d+) minor units$`, w.price)
			sc.Step(`^one public menu publication contains "([^"]*)" at (\d+) minor units in "([^"]*)"$`, func(name string, minor int64, currency string) error {
				if err := w.editions.Succeeded(); err != nil {
					return err
				}
				p := w.editions.Last.Publications
				if len(p) != 1 || p[0].Name != "menu.edition-published" || p[0].Visibility != a.Public {
					return fmt.Errorf("wrong publications: %+v", p)
				}
				event, ok := p[0].Payload.(model.MenuPublished)
				if !ok || event.EditionID != s.Edition || event.Currency != currency || len(event.Offers) != 1 {
					return fmt.Errorf("wrong menu: %+v", event)
				}
				offer := event.Offers[0]
				if offer.Name != name || offer.Minor != minor || offer.Currency != currency || w.editions.Loaded.State.Status != "published" {
					return fmt.Errorf("publication did not freeze the offer: %+v", event)
				}
				return nil
			})
			sc.Step(`^the published offer refers to drink revision (\d+)$`, func(revision uint64) error {
				offer, err := w.offer()
				if err != nil {
					return err
				}
				if offer.DrinkID != s.Drink || offer.DrinkRevision != revision {
					return fmt.Errorf("wrong drink reference: %+v", offer)
				}
				return nil
			})
			sc.Step(`^the drink is renamed to "([^"]*)" and published again$`, func(name string) error {
				_, err := (execution.Bind(s.ForAggregate(&w.drinks, d.RestoreDrink, (*d.Drink).Snapshot, menupub.Drink), func(repository menuports.DrinkWriteRepository) menucommands.ReviseDrinkCommandHandler {
					return menucommands.ReviseDrinkCommandHandler{Repository: repository}
				})).Execute(context.Background(), s.Metadata(s.Drink), menucommands.ReviseDrinkCommand{Name: name})
				if err != nil {
					return err
				}
				if err = w.drinks.Succeeded(); err != nil {
					return err
				}
				return w.publishDrink()
			})
			sc.Step(`^the edition still offers "([^"]*)" at drink revision (\d+)$`, func(name string, revision uint64) error {
				offer, err := w.offer()
				if err != nil {
					return err
				}
				if offer.Name != name || offer.DrinkRevision != revision {
					return fmt.Errorf("edition was rewritten: %+v", offer)
				}
				return nil
			})
			sc.Step(`^the new drink revision produces one private drink publication named "([^"]*)"$`, func(name string) error {
				p := w.drinks.Last.Publications
				if len(p) != 1 || p[0].Name != "menu.drink-published" || p[0].Visibility != a.Private {
					return fmt.Errorf("wrong private mapping: %+v", p)
				}
				event, ok := p[0].Payload.(app.DrinkPublished)
				if !ok || event.DrinkID != s.Drink || event.Revision != 2 || event.Name != name {
					return fmt.Errorf("wrong drink revision: %+v", event)
				}
				return nil
			})
			sc.Step(`^the operator attempts to "([^"]*)" the published edition$`, func(action string) error {
				switch action {
				case "change price":
					return w.price("C1", 350)
				case "add an offer":
					return w.add("C2", 300)
				case "publish again":
					return w.publish()
				}
				return fmt.Errorf("unknown action %q", action)
			})
			sc.Step(`^the draft edition offers "([^"]*)" for (\d+) minor units$`, func(name string, minor int64) error {
				offer, err := w.offer()
				if err != nil {
					return err
				}
				if offer.Name != name || offer.Minor != minor || w.editions.Loaded.State.Status != "draft" {
					return fmt.Errorf("unexpected draft: %+v", w.editions.Loaded)
				}
				return nil
			})
			sc.Step(`^no public menu publication is produced$`, func() error {
				for _, p := range w.editions.Last.Publications {
					if p.Visibility == a.Public {
						return fmt.Errorf("unexpected public event: %+v", p)
					}
				}
				return nil
			})
			sc.Step(`^the drink revision has not reached the menu directory$`, func() { clear(w.directory.Values) })
		}}
	if suite.Run() != 0 {
		t.Fatal("Menu Gherkin scenarios failed")
	}
}
