// Package support owns Fx process lifecycle, cancellation and HTTP resource ordering.
package support

import (
	"context"
	"go.uber.org/fx"
	"net"
	"net/http"
	"sync"
)

// WorkerGroup joins background delivery workers before dependent resources close.
type WorkerGroup struct {
	Context context.Context
	cancel  context.CancelFunc
	workers sync.WaitGroup
}

func NewWorkerGroup(lifecycle fx.Lifecycle) *WorkerGroup {
	ctx, cancel := context.WithCancel(context.Background())
	group := &WorkerGroup{Context: ctx, cancel: cancel}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { group.Close(); return nil }})
	return group
}
func (g *WorkerGroup) Start(run func(context.Context)) {
	g.workers.Add(1)
	go func() { defer g.workers.Done(); run(g.Context) }()
}

// Serve binds synchronously so a failed listener fails application startup.
func Serve(lifecycle fx.Lifecycle, shutdown fx.Shutdowner, server *http.Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
					_ = shutdown.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: server.Shutdown,
	})
}

// Close drains cancelled workers before their dependent resources can close.
func (g *WorkerGroup) Close() { g.cancel(); g.workers.Wait() }
