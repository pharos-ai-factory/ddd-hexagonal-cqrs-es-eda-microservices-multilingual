package postgres

import (
	"encoding/json"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

func TestStoredNestedScalarsRetainPresence(t *testing.T) {
	type price struct {
		Minor int64 `json:"minor"`
	}
	type state struct {
		Offers []price `json:"offers"`
	}
	for _, raw := range []string{`{}`, `null`, `{"offers":[{}]}`, `{"offers":[null]}`,
		`{"offers":[{"minor":null}]}`, `{"offers":[{"minor":"0"}]}`, `{"offers":[{"minor":0.5}]}`} {
		var snapshot state
		if err := decodeStored([]byte(raw), &snapshot); err == nil {
			t.Errorf("accepted incomplete stored price: %s", raw)
		}
	}
	for _, minor := range []int64{0, 300} {
		raw, err := json.Marshal(state{Offers: []price{{Minor: minor}}})
		if err != nil {
			t.Fatal(err)
		}
		var snapshot state
		if err := decodeStored(raw, &snapshot); err != nil || snapshot.Offers[0].Minor != minor {
			t.Fatalf("explicit price lost: %+v %v", snapshot, err)
		}
	}
}

func TestRecordedOutcomesValidateAuthority(t *testing.T) {
	const id = "00000000-0000-4000-8000-000000000001"
	valid := map[string]any{"aggregateId": id, "version": 1, "status": "active"}
	for _, field := range []string{"aggregateId", "version", "status"} {
		for _, value := range []any{nil, true, []any{}} {
			broken := map[string]any{}
			for key, original := range valid {
				broken[key] = original
			}
			broken[field] = value
			raw, _ := json.Marshal(broken)
			if _, err := decodeOutcome(raw, id); err == nil {
				t.Errorf("accepted invalid receipt: %s", raw)
			}
		}
	}
	for _, outcome := range []a.Outcome{
		{AggregateID: id, Version: 1, Status: "active"},
		{AggregateID: id, Rejection: &d.Violation{Code: "not_found", Message: "Missing root"}},
	} {
		raw, _ := json.Marshal(outcome)
		if _, err := decodeOutcome(raw, id); err != nil {
			t.Fatal(err)
		}
		if _, err := decodeOutcome(raw, "00000000-0000-4000-8000-000000000002"); err == nil {
			t.Fatal("accepted another aggregate's receipt")
		}
	}
}
