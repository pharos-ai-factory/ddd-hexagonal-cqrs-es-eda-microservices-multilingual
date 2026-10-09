package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ReviseDrinkCommand expresses the owner use case independently of its transport.
type ReviseDrinkCommand struct {
	Name string `json:"name"`
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
