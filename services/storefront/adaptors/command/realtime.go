package command

import (
	"context"
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// RealtimeBuffer collects explicit publications until the command transaction commits.
type RealtimeBuffer struct {
	active bool
	Events []a.RealtimeMessage
}

func (b *RealtimeBuffer) Publish(ctx context.Context, event a.RealtimeMessage) error {
	if !b.active {
		return fmt.Errorf("realtime bus used outside its transaction")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if (event.Snapshot == nil) == (event.Rejection == nil) {
		return fmt.Errorf("select one realtime payload")
	}
	if event.Target.ID == "" || (event.Target.Kind != "context" && event.Target.Kind != "principal") {
		return fmt.Errorf("select one realtime delivery target")
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	var copy a.RealtimeMessage
	if err = json.Unmarshal(data, &copy); err != nil {
		return err
	}
	b.Events = append(b.Events, copy)
	return nil
}
func (b *RealtimeBuffer) Close() { b.active = false }
