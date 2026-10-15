package workers

import (
	"context"
	"fmt"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	lifecycle "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/runtime"
	"time"
)

func Commands(ctx context.Context, db *postgres.ContextDatabase, url string, decoders map[string]a.Decoder) {
	for ctx.Err() == nil {
		publisher, err := broker.Connect(url)
		if err != nil {
			diagnostics.Record(db.Owner(), "commands.connect", "", "", err, false)
			if !lifecycle.Wait(ctx, time.Second) {
				return
			}
			continue
		}
		for ctx.Err() == nil {
			dispatch, found, err := db.ClaimCommand(ctx, 30*time.Second)
			if err != nil {
				diagnostics.Record(db.Owner(), "commands.claim", "", "", err, false)
				break
			}
			if !found {
				if !lifecycle.Wait(ctx, 100*time.Millisecond) {
					break
				}
				continue
			}
			var message a.Message
			decode, ok := decoders[dispatch.Name]
			if !ok {
				err = fmt.Errorf("unregistered owner command codec")
			} else {
				message, err = decode(dispatch.Body)
			}
			if err == nil && (message.ID != dispatch.EventID || message.Context != db.Owner() || message.Name != dispatch.Name+".command" || message.Visibility != "command") {
				err = fmt.Errorf("stored command metadata mismatch")
			}
			if err == nil {
				publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = publisher.Publish(publishCtx, message, dispatch.Body)
				cancel()
			}
			if err != nil {
				_, retryErr := db.RetryCommand(ctx, dispatch, diagnostics.Class(err))
				if retryErr != nil {
					diagnostics.Record(db.Owner(), "commands.retry", dispatch.EventID, message.CorrelationID, retryErr, false)
				}
				diagnostics.Record(db.Owner(), "commands.publish", dispatch.EventID, message.CorrelationID, err, false)
				break
			}
			completed, err := db.CompleteCommand(ctx, dispatch)
			if err != nil || !completed {
				diagnostics.Record(db.Owner(), "commands.complete", dispatch.EventID, message.CorrelationID, err, false)
			}
		}
		publisher.Close()
		if !lifecycle.Wait(ctx, time.Second) {
			return
		}
	}
}
