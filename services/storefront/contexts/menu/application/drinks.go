package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateDrinkCommand expresses the owner CreateDrink use case independently of transport.
type CreateDrinkCommand struct {
	Name string `json:"name"`
}

// CreateDrinkCommandHandler applies CreateDrink through one aggregate command transaction.
type CreateDrinkCommandHandler struct {
	Drinks a.AggregateCommandPort[d.DrinkState]
}

func (h CreateDrinkCommandHandler) Execute(ctx context.Context, m a.Metadata, c CreateDrinkCommand) (a.Outcome, error) {
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

// PublishDrinkCommandHandler applies PublishDrink through one aggregate command transaction.
type PublishDrinkCommandHandler struct {
	Drinks a.AggregateCommandPort[d.DrinkState]
}

func (h PublishDrinkCommandHandler) Execute(ctx context.Context, m a.Metadata, _ PublishDrinkCommand) (a.Outcome, error) {
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
				result.Publications = append(result.Publications, a.Publication{Name: "menu.drink-published", Visibility: a.Private, Payload: DrinkPublished{DrinkID: s.ID, Name: s.Name, Revision: s.Revision}})
			}
		}
		return result, nil
	})
}

// ReviseDrinkCommandHandler applies ReviseDrink through one aggregate command transaction.
type ReviseDrinkCommandHandler struct {
	Drinks a.AggregateCommandPort[d.DrinkState]
}

func (h ReviseDrinkCommandHandler) Execute(ctx context.Context, m a.Metadata, c ReviseDrinkCommand) (a.Outcome, error) {
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

// PublishDrinkCommand expresses the owner use case independently of its transport.
type PublishDrinkCommand struct{}

// ReviseDrinkCommand expresses the owner use case independently of its transport.
type ReviseDrinkCommand struct {
	Name string `json:"name"`
}
