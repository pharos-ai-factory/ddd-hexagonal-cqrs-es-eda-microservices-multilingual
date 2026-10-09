package main

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	"go.uber.org/fx"
)

func composition() fx.Option {
	return fx.Options(fx.Provide(func() (*support.StorefrontRuntime, error) { return support.Open("menu", "ordering") }),
		menuModule(), orderingModule(), fx.Invoke(startService))
}
func main() { fx.New(composition()).Run() }

func startService(lifecycle fx.Lifecycle, shutdown fx.Shutdowner, service *support.StorefrontRuntime) {
	done := make(chan error, 1)
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				err := service.Run()
				done <- err
				if err != nil {
					_ = shutdown.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			service.Stop()
			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
