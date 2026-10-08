package backend

import (
	"context"
	"errors"
	test "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/internal/testsupport"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/menu"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	"net/http"
	"testing"
)

const id = "11111111-1111-4111-8111-111111111111"

type caller func(context.Context, pb.Request) (pb.Reply, error)

func (f caller) Call(ctx context.Context, r pb.Request) (pb.Reply, error) { return f(ctx, r) }
func TestAPIAuthenticatesAndTranslatesIntoCommandContract(t *testing.T) {
	var captured pb.Request
	c := config()
	c.Owners = []string{"menu"}
	c.Requests = caller(func(_ context.Context, r pb.Request) (pb.Reply, error) {
		captured = r
		return &menu.Reply{ContractVersion: 1, RequestId: r.GetRequestId(), Context: "menu", Payload: &menu.Reply_Outcome{Outcome: &pb.Outcome{AggregateId: id, Version: 1, Status: "draft"}}}, nil
	})
	mux := backendHandler(c)
	if w := request(t, mux, "GET", "/api/v1/menu/drinks", "", nil); w.Code != 401 {
		t.Fatal("anonymous API access accepted")
	}
	path := "/api/v1/menu/drinks/" + id
	if w := request(t, mux, "POST", path, `{"name":"Coffee"}`, map[string]string{"Cookie": "cafe_session=valid", "Origin": "http://attacker.local"}); w.Code != 403 || captured != nil {
		t.Fatal("foreign origin mutation accepted")
	}
	w := request(t, mux, "POST", path, `{"name":"Coffee"}`, map[string]string{"Cookie": "cafe_session=valid", "Origin": "http://cafe.local", "Authorization": "Bearer attacker", "Idempotency-Key": id, "If-Match": "2"})
	if w.Code != 200 || captured == nil {
		t.Fatalf("command unavailable %d", w.Code)
	}
	command := captured.(*menu.Request).GetCommand()
	if command.GetCreateDrink().GetName() != "Coffee" || command.Metadata.CommandId != id || command.Metadata.GetExpectedVersion() != 2 || command.Metadata.CorrelationId != id || captured.GetRequestId() == id {
		t.Fatal("command translation changed identity, input or version")
	}
	if object := rpc.Object(command); len(object) != 2 {
		t.Fatal("browser transport details escaped boundary")
	}
}
func config() Config {
	return Config{Sessions: test.Sessions{}, CLIKey: "cli", Origins: []string{"http://cafe.local"}}
}

var request = test.Request
var unavailable = errors.New("broker credential: private-secret")

func backendHandler(c Config) http.Handler {
	mux := contract.NewMux("api", map[string]bool{"menu": true})
	for _, pattern := range mux.Patterns() {
		if mux.Owner(pattern) == "" {
			mux.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(500) })
		}
	}
	c.Mount(mux)
	return mux.Handler()
}
