package support

import (
	"context"
	"go.uber.org/fx"
	"testing"
)

func TestWorkerDrainsBeforeStopReturns(t *testing.T) {
	started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	app := fx.New(fx.NopLogger, fx.Provide(NewWorkerGroup), fx.Invoke(func(g *WorkerGroup, lc fx.Lifecycle) {
		lc.Append(fx.Hook{OnStart: func(context.Context) error {
			g.Start(func(ctx context.Context) { close(started); <-ctx.Done(); <-release; close(finished) })
			return nil
		}})
	}))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	<-started
	stopped := make(chan error, 1)
	go func() { stopped <- app.Stop(t.Context()) }()
	select {
	case <-stopped:
		t.Fatal("returned before in-flight work completed")
	default:
	}
	close(release)
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("work not drained")
	}
}
