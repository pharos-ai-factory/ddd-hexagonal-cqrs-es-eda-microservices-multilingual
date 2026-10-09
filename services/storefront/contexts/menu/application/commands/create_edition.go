package commands

import (
	"context"
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
	Editions a.AggregateCommandPort[d.EditionState]
}

func (h CreateEditionCommandHandler) Execute(ctx context.Context, m a.Metadata, c CreateEditionCommand) (a.Outcome, error) {
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("already_exists", "The edition already exists")
		}
		edition, err := d.NewEdition(m.AggregateID, c.Currency)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		return a.Changed(edition.Snapshot(), "draft"), nil
	})
}
