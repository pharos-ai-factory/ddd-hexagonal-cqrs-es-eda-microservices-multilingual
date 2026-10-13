// Package di assembles Storefront's resources, context modules and lifecycle.
package di

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	"go.uber.org/fx"
)

// Module supplies the service graph consumed by support.Run.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(func(scope *support.ResourceScope) (*support.StorefrontRuntime, error) {
			service, err := support.Open("menu", "ordering")
			if err == nil {
				scope.Add(service.Close)
			}
			return service, err
		}),
		menuModule(),
		orderingModule(),
		fx.Invoke(startService),
	)
}
