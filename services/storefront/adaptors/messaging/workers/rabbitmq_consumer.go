package workers

import (
	"context"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	lifecycle "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/runtime"
	"time"
)

func Consumer(ctx context.Context, url string, sub broker.Subscription, decode a.Decoder, ids a.IDFactory) {
	for ctx.Err() == nil {
		err := broker.Consume(ctx, url, sub, decode, ids)
		if ctx.Err() != nil {
			return
		}
		diagnostics.Record(sub.Binding.Context, sub.Binding.Consumer, "", "", err, false)
		if !lifecycle.Wait(ctx, time.Second) {
			return
		}
	}
}
