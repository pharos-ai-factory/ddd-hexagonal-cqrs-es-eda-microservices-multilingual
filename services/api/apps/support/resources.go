package support

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/fx"
	"sync"
	"time"
)

// ResourceScope closes acquired resources even when graph construction fails before Fx starts.
type ResourceScope struct {
	once  sync.Once
	close []func()
}

func (s *ResourceScope) Add(close func()) { s.close = append(s.close, close) }
func (s *ResourceScope) Close() {
	s.once.Do(func() {
		for i := len(s.close) - 1; i >= 0; i-- {
			s.close[i]()
		}
	})
}

// Run preserves cleanup on composition, startup and shutdown failures.
func Run(options ...fx.Option) error {
	scope := &ResourceScope{}
	defer scope.Close()
	app := fx.New(append(options, fx.Supply(scope))...)
	if err := app.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		return err
	}
	signal := <-app.Wait()
	stopped, done := context.WithTimeout(context.Background(), 30*time.Second)
	defer done()
	err := app.Stop(stopped)
	if signal.ExitCode != 0 {
		err = errors.Join(err, fmt.Errorf("application requested exit code %d", signal.ExitCode))
	}
	return err
}
