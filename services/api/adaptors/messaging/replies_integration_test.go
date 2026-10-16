//go:build integration

package messaging

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
)

func liveClient(t *testing.T) *RabbitMQRequestClient {
	t.Helper()
	url := os.Getenv("API_BROKER_URL")
	if url == "" || !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Skip("requires disposable stack")
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := NewClient()
	done := make(chan struct{})
	go func() { c.Run(ctx, url, []string{"menu", "collection", "loyalty"}); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		ready := c.channel != nil
		c.mu.Unlock()
		if ready {
			return c
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("API reply connection unavailable")
	return nil
}

// Exercise real query replies and committed command replies in Go, Python and TypeScript.
func exerciseClients(t *testing.T, clients ...*RabbitMQRequestClient) {
	t.Helper()
	var group sync.WaitGroup
	for i := range 24 {
		group.Add(1)
		go func() {
			defer group.Done()
			owners := []string{"menu", "collection", "loyalty"}
			owner := owners[i%3]
			request, err := pb.NewRequest(owner, newConnectionID())
			if err != nil {
				t.Error(err)
				return
			}
			kind := "query"
			name := map[string]string{"menu": "listDrinks", "collection": "listPickups", "loyalty": "listRewards"}[owner]
			payload := map[string]any{}
			if i%2 == 0 {
				kind = "command"
				name = map[string]string{"menu": "createDrink", "collection": "collectOrder", "loyalty": "redeemReward"}[owner]
				input := map[string]any{"name": "Reply isolation"}
				if owner == "collection" {
					input = map[string]any{"code": "1234"}
				}
				if owner == "loyalty" {
					input = map[string]any{"orderId": newConnectionID()}
				}
				payload["metadata"] = map[string]any{"commandId": newConnectionID(), "aggregateId": newConnectionID(), "expectedVersion": 0, "correlationId": request.GetRequestId(), "actorId": newConnectionID()}
				payload[name] = input
			} else {
				payload[name] = map[string]any{}
			}
			outer, err := NewPayload(request, kind)
			if err != nil {
				t.Error(err)
				return
			}
			if err = FromObject(payload, outer); err != nil {
				t.Error(err)
				return
			}
			if err = SetPayload(request, kind, outer); err != nil {
				t.Error(err)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			reply, err := clients[i%len(clients)].Call(ctx, request)
			if err != nil {
				t.Error(fmt.Errorf("%s %s: %w", owner, kind, err))
				return
			}
			if reply.GetRequestId() != request.GetRequestId() || reply.GetContext() != owner || reply.GetError() != nil || (kind == "command" && reply.GetOutcome() == nil) {
				t.Errorf("invalid %s reply: %v", owner, reply)
			}
		}()
	}
	group.Wait()
}

func TestOverlappingAPIConnectionsAndReconnect(t *testing.T) {
	first, second := liveClient(t), liveClient(t)
	exerciseClients(t, first, second)
	first.mu.Lock()
	old, channel := first.routes["menu"], first.channel
	first.mu.Unlock()
	second.mu.Lock()
	other := second.routes["menu"]
	second.mu.Unlock()
	if old == other {
		t.Fatal("API instances share reply queues")
	}
	if err := channel.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		first.mu.Lock()
		ready, route := first.channel != nil, first.routes["menu"]
		first.mu.Unlock()
		if ready && route != old {
			exerciseClients(t, first, second)
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("API reconnect did not replace its reply queues")
}
