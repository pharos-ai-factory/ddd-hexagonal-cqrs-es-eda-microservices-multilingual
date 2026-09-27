package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

type CreateDrink struct {
	Name string `json:"name"`
}
type CreateDrinkHandler struct{ Drinks a.CommandPort[d.DrinkState] }

func (h CreateDrinkHandler) Execute(ctx context.Context, m a.Metadata, c CreateDrink) (a.Outcome, error) {
	return h.Drinks.Execute(ctx, m, func(s a.Loaded[d.DrinkState]) (a.Mutation[d.DrinkState], error) {
		if s.Exists {
			return a.Mutation[d.DrinkState]{}, core.Reject("already_exists", "The drink already exists")
		}
		drink, err := d.NewDrink(m.AggregateID, c.Name)
		if err != nil {
			return a.Mutation[d.DrinkState]{}, err
		}
		return a.Changed(drink.Snapshot(), "draft"), nil
	})
}

type PublishDrinkHandler struct{ Drinks a.CommandPort[d.DrinkState] }

func (h PublishDrinkHandler) Execute(ctx context.Context, m a.Metadata, _ struct{}) (a.Outcome, error) {
	return h.Drinks.Execute(ctx, m, func(s a.Loaded[d.DrinkState]) (a.Mutation[d.DrinkState], error) {
		if !s.Exists {
			return a.Mutation[d.DrinkState]{}, core.Reject("not_found", "The drink does not exist")
		}
		drink, err := d.RestoreDrink(s.State)
		if err != nil {
			return a.Mutation[d.DrinkState]{}, err
		}
		if err = drink.Publish(); err != nil {
			return a.Mutation[d.DrinkState]{}, err
		}
		result := a.Changed(drink.Snapshot(), "published")
		for _, fact := range drink.Events() {
			if fact.Name == "DrinkPublished" {
				s := fact.Data.(d.DrinkState)
				result.Publications = append(result.Publications, a.Publication{Name: "menu.drink-published", Visibility: a.Private, Payload: model.DrinkPublished{DrinkID: s.ID, Name: s.Name, Revision: s.Revision}})
			}
		}
		return result, nil
	})
}

type ReviseDrinkHandler struct{ Drinks a.CommandPort[d.DrinkState] }

func (h ReviseDrinkHandler) Execute(ctx context.Context, m a.Metadata, c CreateDrink) (a.Outcome, error) {
	return h.Drinks.Execute(ctx, m, func(s a.Loaded[d.DrinkState]) (a.Mutation[d.DrinkState], error) {
		if !s.Exists {
			return a.Mutation[d.DrinkState]{}, core.Reject("not_found", "The drink does not exist")
		}
		drink, err := d.RestoreDrink(s.State)
		if err != nil {
			return a.Mutation[d.DrinkState]{}, err
		}
		if err = drink.Revise(c.Name); err != nil {
			return a.Mutation[d.DrinkState]{}, err
		}
		status := "draft"
		if drink.Snapshot().Published {
			status = "published"
		}
		return a.Mutation[d.DrinkState]{State: drink.Snapshot(), Changed: len(drink.Events()) > 0, Status: status}, nil
	})
}
