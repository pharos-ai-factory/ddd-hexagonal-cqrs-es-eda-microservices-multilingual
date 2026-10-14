//go:build integration

package postgres

import (
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	s "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

func snapshotDecisions[S any](db *ContextDatabase, kind string) s.SnapshotDecisionFixture[S] {
	return s.SnapshotDecisionFixture[S]{Transaction: NewAggregateTransaction[S, S](db, kind, func(repo a.WriteRepository[S]) a.WriteRepository[S] { return repo })}
}
