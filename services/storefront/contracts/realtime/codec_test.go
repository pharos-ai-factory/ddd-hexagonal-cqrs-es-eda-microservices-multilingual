package realtime

import (
	"testing"

	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime/generated/cafe/realtime/v1"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"google.golang.org/protobuf/proto"
)

func TestEncodePreservesApplicationDeliveryTargets(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	state, err := Encode(id, "ordering", "order", id, 1, a.SnapshotMessage(a.ContextTarget("ordering"), map[string]any{"id": id}))
	if err != nil {
		t.Fatal(err)
	}
	var publication pb.Publication
	if err = proto.Unmarshal(state, &publication); err != nil || publication.GetTarget().GetContext() != "ordering" {
		t.Fatal("context target was not encoded")
	}
	rejection := a.RejectionMessage(a.PrincipalTarget(id), d.Violation{Code: "not_found", Message: "missing"})
	rejection.CommandID, rejection.CorrelationID = id, id
	body, err := Encode(id, "ordering", "order", id, 0, rejection)
	if err != nil {
		t.Fatal(err)
	}
	if err = proto.Unmarshal(body, &publication); err != nil || publication.GetTarget().GetPrincipal() != id {
		t.Fatal("principal target was not encoded")
	}
}
