package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ReviseDrinkCommandHandler applies ReviseDrink through one aggregate command transaction.
type ReviseDrinkCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.DrinkWriteRepository
}

func (h ReviseDrinkCommandHandler) Execute(ctx context.Context, m a.CommandContext, c menucontracts.ReviseDrinkCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, core.Reject("not_found", "The drink does not exist")
	}
	drink := s.State
	if err = drink.Revise(c.Name); err != nil {
		return a.CommandResult{}, err
	}
	status := "draft"
	if drink.Snapshot().Published {
		status = "published"
	}
	if len(drink.Events()) > 0 {
		if err = h.Repository.Save(ctx, drink); err != nil {
			return a.CommandResult{}, err
		}
		if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: drink.Snapshot()}); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result(status), nil
}
