package support

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// PublicationProbe records application mapping without claiming transaction or delivery evidence.
type PublicationProbe[P any] struct {
	Payload P
	Calls   int
}

func (p *PublicationProbe[P]) Publish(_ context.Context, m a.Metadata, payload P) (a.Outcome, error) {
	p.Payload = payload
	p.Calls++
	return a.Outcome{AggregateID: m.AggregateID, Version: 1, Status: "publication_queued"}, nil
}
