package postgres

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// Persisted scalar fields must be explicit: a missing price is not a free price.
// Domain restoration still owns the business invariants of the decoded snapshot.
func decodeStored[S any](data []byte, state *S) error {
	if err := requiredFields(data, reflect.TypeFor[S]()); err != nil {
		return d.Corrupt(err.Error())
	}
	if err := json.Unmarshal(data, state); err != nil {
		return d.Corrupt(err.Error())
	}
	return nil
}

func requiredFields(data []byte, shape reflect.Type) error {
	if shape.Kind() == reflect.Pointer {
		shape = shape.Elem()
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		if shape.Kind() == reflect.Slice || shape.Kind() == reflect.Map {
			return nil // Nil collections are written as null by encoding/json.
		}
		return fmt.Errorf("null %s", shape)
	}
	switch shape.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		for i := 0; i < shape.NumField(); i++ {
			field := shape.Field(i)
			name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
			if !field.IsExported() || name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			value, exists := fields[name]
			if !exists && strings.Contains(options, "omitempty") {
				continue
			}
			if !exists {
				return fmt.Errorf("missing stored field %s", name)
			}
			if err := requiredFields(value, field.Type); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	case reflect.Slice, reflect.Array:
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for _, value := range values {
			if err := requiredFields(value, shape.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkRootIdentity(data []byte, id string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return d.Corrupt(err.Error())
	}
	// Generic technical fixtures may have no identity field.
	if value, exists := fields["id"]; exists {
		var stored string
		if json.Unmarshal(value, &stored) != nil || stored != id {
			return d.Corrupt("snapshot identity differs from its storage key")
		}
	}
	return nil
}

func decodeOutcome(data []byte, id string) (a.Outcome, error) {
	var result a.Outcome
	if err := decodeStored(data, &result); err != nil {
		return a.Outcome{}, err
	}
	if result.AggregateID != id || (result.Rejection == nil && result.Status == "") ||
		(result.Rejection != nil && (result.Rejection.Code == "" || result.Rejection.Message == "")) {
		return a.Outcome{}, d.Corrupt("invalid recorded outcome")
	}
	return result, nil
}
