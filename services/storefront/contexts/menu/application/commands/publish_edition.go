package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// PublishEditionCommandHandler applies PublishEdition through one aggregate command transaction.
type PublishEditionCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeMessage]
	Repository ports.EditionWriteRepository
}

func (h PublishEditionCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ menucontracts.PublishEditionCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, core.Reject("not_found", "The edition does not exist")
	}
	edition := s.State
	if err = edition.Publish(); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, edition); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Realtime.Publish(ctx, a.SnapshotMessage(a.ContextTarget("menu"), edition.Snapshot())); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("published"), nil
}
