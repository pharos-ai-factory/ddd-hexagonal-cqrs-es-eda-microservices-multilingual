package realtime

import (
	"encoding/json"
	"fmt"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime/generated/cafe/realtime/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Encode(eventID, owner, kind, aggregateID string, revision uint64, snapshot []byte) ([]byte, error) {
	if !((owner == "menu" && (kind == "drink" || kind == "edition")) || (owner == "ordering" && kind == "order")) {
		return nil, fmt.Errorf("unregistered Storefront browser projection")
	}
	raw, err := json.Marshal(map[string]any{"eventId": eventID, "contractVersion": 1, "context": owner,
		"aggregateKind": kind, "aggregateId": aggregateID, "revision": revision, kind: json.RawMessage(snapshot)})
	if err != nil {
		return nil, err
	}
	var publication pb.Publication
	if err = protojson.Unmarshal(raw, &publication); err != nil {
		return nil, err
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(&publication)
}
