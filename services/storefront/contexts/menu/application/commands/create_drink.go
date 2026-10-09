package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
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
	Repository ports.DrinkWriteRepository
}

func (h CreateDrinkCommandHandler) Execute(ctx context.Context, m a.CommandContext, c CreateDrinkCommand) (a.CommandResult, error) {
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
	return a.Result("draft"), nil
}
