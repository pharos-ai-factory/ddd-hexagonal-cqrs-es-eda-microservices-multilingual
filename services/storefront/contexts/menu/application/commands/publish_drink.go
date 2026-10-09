package commands

import (
	"context"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// PublishDrinkCommand expresses the owner use case independently of its transport.
type PublishDrinkCommand struct{}

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
				result.Publications = append(result.Publications, a.Publication{Name: "menu.drink-published", Visibility: a.Private, Payload: app.DrinkPublished{DrinkID: s.ID, Name: s.Name, Revision: s.Revision}})
			}
		}
		return result, nil
	})
}
