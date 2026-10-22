package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// PublishDrinkCommandHandler applies PublishDrink through one aggregate command transaction.
type PublishDrinkCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeMessage]
	Repository ports.DrinkWriteRepository
}

func (h PublishDrinkCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ menucontracts.PublishDrinkCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, core.Reject("not_found", "The drink does not exist")
	}
	drink := s.State
	if err = drink.Publish(); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, drink); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Realtime.Publish(ctx, a.SnapshotMessage(a.ContextTarget("menu"), drink.Snapshot())); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("published"), nil
}
