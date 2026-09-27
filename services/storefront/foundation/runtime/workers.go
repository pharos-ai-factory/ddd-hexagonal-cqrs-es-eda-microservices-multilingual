package runtime

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"log/slog"
	"time"
)

func Wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
func Relay(ctx context.Context, db *postgres.Database, url string, decode a.Decoder) {
	for ctx.Err() == nil {
		publisher, err := broker.Connect(url)
		if err != nil {
			if !Wait(ctx, time.Second) {
				return
			}
			continue
		}
		for ctx.Err() == nil {
			dispatch, found, err := db.Claim(ctx, 30*time.Second)
			if err != nil {
				slog.Error("dispatch claim failed", "error", err)
				break
			}
			if !found {
				if !Wait(ctx, 100*time.Millisecond) {
					break
				}
				continue
			}
			message, err := decode(dispatch.Body)
			if err == nil {
				publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = publisher.Publish(publishCtx, message, dispatch.Body)
				cancel()
			}
			if err != nil {
				_, _ = db.Retry(ctx, dispatch, err.Error())
				slog.Warn("dispatch retained", "event", dispatch.EventID, "error", err)
				break
			}
			completed, err := db.Complete(ctx, dispatch)
			if err != nil || !completed {
				slog.Warn("dispatch completion requires recovery", "event", dispatch.EventID, "error", err)
			}
		}
		publisher.Close()
		if !Wait(ctx, time.Second) {
			return
		}
	}
}
func Consumer(ctx context.Context, url string, sub broker.Subscription, decode a.Decoder, ids a.IDFactory) {
	for ctx.Err() == nil {
		err := broker.Consume(ctx, url, sub, decode, ids)
		if ctx.Err() != nil {
			return
		}
		slog.Warn("consumer reconnecting", "consumer", sub.Binding.Consumer, "error", err)
		if !Wait(ctx, time.Second) {
			return
		}
	}
}
