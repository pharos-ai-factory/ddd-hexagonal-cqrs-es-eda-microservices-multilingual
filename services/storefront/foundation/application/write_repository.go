package application

import "context"

// CommandResult is the business status returned by a feature command handler.
type CommandResult struct {
	Status string
}

// WriteRepository loads and saves one targeted aggregate within an active command transaction.
type WriteRepository[A any] interface {
	Get(context.Context, string) (Loaded[A], error)
	Save(context.Context, A) error
}

// Result selects the business status for a successful command.
func Result(status string) CommandResult { return CommandResult{Status: status} }

// CommandContext identifies the aggregate targeted by a business operation.
type CommandContext struct{ Target string }
