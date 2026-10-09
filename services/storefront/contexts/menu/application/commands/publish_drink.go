package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// PublishDrinkCommand expresses the owner use case independently of its transport.
type PublishDrinkCommand struct{}

// PublishDrinkCommandHandler applies PublishDrink through one aggregate command transaction.
type PublishDrinkCommandHandler struct {
	Repository ports.DrinkWriteRepository
}

func (h PublishDrinkCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ PublishDrinkCommand) (a.CommandResult, error) {
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
	return a.Result("published"), nil
}
