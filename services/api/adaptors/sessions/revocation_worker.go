package sessions

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// Run retries durable session disconnect work after the connect-grant fence.
func (s *ValkeySessionStore) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := s.client.ZRangeArgs(ctx, redis.ZRangeArgs{Key: prefix + "expirations",
				Start: "-inf", Stop: fmt.Sprint(time.Now().Unix()), ByScore: true, Count: 100}).Result()
			if err != nil {
				s.record("scan", err)
				continue
			}
			for _, key := range expired {
				_, expireErr := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
					pipe.Del(ctx, prefix+"session:"+key)
					pipe.ZRem(ctx, prefix+"expirations", key)
					pipe.SAdd(ctx, prefix+"disconnects", a.OperatorID+"|"+key)
					return nil
				})
				s.record("expire", expireErr)
			}
			actors, err := s.client.SMembers(ctx, prefix+"disconnects").Result()
			if err != nil {
				s.record("scan", err)
				continue
			}
			for _, work := range actors {
				// Each revoked session has its own work identity. Completing an
				// older disconnect cannot remove a concurrent revocation.
				// Start the fence only after observing durable revocation. Every
				// earlier connect grant then expires before the final disconnect.
				fence := prefix + "fence:" + work
				if err := s.client.SetNX(ctx, fence, time.Now().Add(a.ConnectLifetime).Unix(), Lifetime).Err(); err != nil {
					s.record("fence", err)
					continue
				}
				deadline, err := s.client.Get(ctx, fence).Int64()
				s.record("fence.read", err)
				// Centrifugo accepts expire_at == now, so cross the whole second.
				if err != nil || time.Now().Unix() <= deadline {
					continue
				}
				actor, _, _ := strings.Cut(work, "|")
				if err := s.disconnect(ctx, actor); err == nil {
					s.complete(ctx, work, fence)
				} else {
					s.record("disconnect", err)
				}
			}
		}
	}
}

// Keep the pending evidence when completion fails; a later worker will retry.
func (s *ValkeySessionStore) complete(ctx context.Context, work, fence string) {
	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, prefix+"disconnects", work)
		pipe.Del(ctx, fence)
		return nil
	})
	s.record("complete", err)
}
