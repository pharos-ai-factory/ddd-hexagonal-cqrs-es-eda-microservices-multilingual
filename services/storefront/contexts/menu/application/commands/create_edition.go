package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateEditionCommand expresses the owner CreateEdition use case independently of transport.
type CreateEditionCommand struct {
	Currency string `json:"currency"`
}

// CreateEditionCommandHandler applies CreateEdition through one aggregate command transaction.
type CreateEditionCommandHandler struct {
	Repository ports.EditionWriteRepository
}

func (h CreateEditionCommandHandler) Execute(ctx context.Context, m a.CommandContext, c CreateEditionCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if s.Exists {
		return a.CommandResult{}, core.Reject("already_exists", "The edition already exists")
	}
	edition, err := d.NewEdition(m.Target, c.Currency)
	if err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, edition); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("draft"), nil
}
