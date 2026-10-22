package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateDrinkCommandHandler applies CreateDrink through one aggregate command transaction.
type CreateDrinkCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeMessage]
	Repository ports.DrinkWriteRepository
}

func (h CreateDrinkCommandHandler) Execute(ctx context.Context, m a.CommandContext, c menucontracts.CreateDrinkCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if s.Exists {
		return a.CommandResult{}, core.Reject("already_exists", "The drink already exists")
	}
	drink, err := d.NewDrink(m.Target, c.Name)
	if err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, drink); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Realtime.Publish(ctx, a.SnapshotMessage(a.ContextTarget("menu"), drink.Snapshot())); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("draft"), nil
}
