// Package envelope validates technical publication identities.
package envelope

import (
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ValidateMetadata checks identifiers and portable source revisions.
func ValidateMetadata(m a.Message) error {
	for _, id := range []string{m.ID, m.AggregateID, m.CorrelationID, m.CausationID} {
		if err := core.ValidateID(id); err != nil {
			return err
		}
	}
	if m.AggregateVersion == 0 || m.AggregateVersion > 9007199254740991 || m.OccurredAt.IsZero() {
		return fmt.Errorf("incomplete aggregate metadata")
	}
	return nil
}
