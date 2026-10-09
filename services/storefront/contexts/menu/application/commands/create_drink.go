package commands

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
