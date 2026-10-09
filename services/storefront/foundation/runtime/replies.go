package runtime

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"time"
)

func Replies(ctx context.Context, db *postgres.ContextDatabase, url, owner string) {
	for ctx.Err() == nil {
		publisher, err := broker.Connect(url)
		if err == nil {
			for ctx.Err() == nil {
				row, found, claimErr := db.ClaimReply(ctx)
				if claimErr != nil {
					err = claimErr
					break
				}
				if !found {
					select {
					case <-ctx.Done():
					case <-time.After(postgres.ReplyPollInterval):
					}
					continue
				}
				reason := ""
				if row.Expired {
					reason = "expired"
				} else {
					publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					err = publisher.PublishReply(publishCtx, owner, row.EventID, row.Body)
					cancel()
					if err != nil {
						reason = "publication_failed"
					}
				}
				if finishErr := db.FinishReply(ctx, row, reason); finishErr != nil {
					err = finishErr
				}
				if err != nil {
					break
				}
			}
			publisher.Close()
		}
		if ctx.Err() != nil {
			return
		}
		diagnostics.Record(owner, "replies.reconnect", "", "", err, false)
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
}
