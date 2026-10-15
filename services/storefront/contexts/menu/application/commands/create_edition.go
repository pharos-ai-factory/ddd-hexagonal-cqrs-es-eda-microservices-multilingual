package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateEditionCommandHandler applies CreateEdition through one aggregate command transaction.
type CreateEditionCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.EditionWriteRepository
}

func (h CreateEditionCommandHandler) Execute(ctx context.Context, m a.CommandContext, c menucontracts.CreateEditionCommand) (a.CommandResult, error) {
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
	if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: edition.Snapshot()}); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("draft"), nil
}
