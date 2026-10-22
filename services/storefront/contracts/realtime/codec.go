package realtime

import (
	"encoding/json"
	"fmt"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime/generated/cafe/realtime/v1"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Encode(eventID, owner, kind, aggregateID string, revision uint64, event a.RealtimeMessage) ([]byte, error) {
	if !((owner == "menu" && (kind == "drink" || kind == "edition")) || (owner == "ordering" && kind == "order")) {
		return nil, fmt.Errorf("unregistered Storefront realtime message")
	}
	field, payload := kind, event.Snapshot
	if event.Rejection != nil {
		field, payload = "operationRejected", map[string]any{"commandId": event.CommandID, "correlationId": event.CorrelationID, "code": event.Rejection.Code, "message": event.Rejection.Message}
	}
	if event.Target.ID == "" || (event.Target.Kind != "context" && event.Target.Kind != "principal") {
		return nil, fmt.Errorf("invalid realtime delivery target")
	}
	raw, err := json.Marshal(map[string]any{"eventId": eventID, "contractVersion": 1, "context": owner,
		"aggregateKind": kind, "aggregateId": aggregateID, "revision": revision,
		"target": map[string]string{string(event.Target.Kind): event.Target.ID}, field: payload})
	if err != nil {
		return nil, err
	}
	var publication pb.Publication
	if err = protojson.Unmarshal(raw, &publication); err != nil {
		return nil, err
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(&publication)
}
