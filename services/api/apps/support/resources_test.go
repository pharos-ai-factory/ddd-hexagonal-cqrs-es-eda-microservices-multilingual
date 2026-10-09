package support

import (
	"context"
	"errors"
	"go.uber.org/fx"
	"testing"
)

func TestScopeClosesAfterGraphFailure(t *testing.T) {
	closed := 0
	err := Run(fx.NopLogger, fx.Provide(func(scope *ResourceScope) (*int, error) {
		scope.Add(func() { closed++ })
		return nil, errors.New("composition")
	}), fx.Invoke(func(*int) {}))
	if err == nil || closed != 1 {
		t.Fatalf("error=%v closed=%d", err, closed)
	}
}
func TestPartialStartRollsBackBeforeResourceDisposal(t *testing.T) {
	order := []string{}
	err := Run(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle, scope *ResourceScope) {
		scope.Add(func() { order = append(order, "resource") })
		lc.Append(fx.Hook{OnStart: func(context.Context) error { order = append(order, "start"); return nil }, OnStop: func(context.Context) error { order = append(order, "worker"); return nil }})
		lc.Append(fx.Hook{OnStart: func(context.Context) error { return errors.New("listener") }})
	}))
	if err == nil || len(order) != 3 || order[1] != "worker" || order[2] != "resource" {
		t.Fatalf("%v %v", err, order)
	}
}

func TestRequestedFailureExitIsPreserved(t *testing.T) {
	err := Run(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle, shutdown fx.Shutdowner) {
		lc.Append(fx.Hook{OnStart: func(context.Context) error { return shutdown.Shutdown(fx.ExitCode(1)) }})
	}))
	if err == nil {
		t.Fatal("failure shutdown was reported as success")
	}
}
