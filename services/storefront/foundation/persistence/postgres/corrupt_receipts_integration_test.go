//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

func TestCorruptReceiptsRequireRepairBeforeRetry(t *testing.T) {
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Fatal("receipt corruption requires the disposable integration project")
	}
	db, owner := fixtureDB(t), admin(t)
	for _, receipt := range []string{"command", "consumer"} {
		for _, shape := range []string{"empty", "null", "incomplete_rejection"} {
			t.Run(receipt+"/"+shape, func(t *testing.T) {
				id := NewID()
				clean(t, owner, id)
				m := metadata(id, 0)
				m.Consumer, m.SourceEventID, m.SourceHash = "ordering.test-receipt-"+id, NewID(), "original-wire"
				t.Cleanup(func() {
					if _, err := owner.Exec(context.Background(), `DELETE FROM cafe.consumer_receipts WHERE consumer=$1`, m.Consumer); err != nil {
						t.Error(err)
					}
				})
				initial := m
				if receipt == "command" {
					initial.Consumer = ""
				}
				store := Command[counter](db, "test_counter")
				first, err := store.Execute(t.Context(), initial, increment)
				if err != nil || first.Rejection != nil {
					t.Fatalf("fixture creation: %+v %v", first, err)
				}
				original, err := json.Marshal(first)
				if err != nil {
					t.Fatal(err)
				}
				broken := []byte(`{}`)
				if shape == "null" {
					broken = []byte(`null`)
				} else if shape == "incomplete_rejection" {
					broken, err = json.Marshal(map[string]any{"aggregateId": id, "version": 1, "status": "",
						"rejection": map[string]string{"code": "not_found"}})
					if err != nil {
						t.Fatal(err)
					}
				}
				table, key := "command_receipts", "command_id"
				identity := m.ID
				if receipt == "consumer" {
					table, key, identity = "consumer_receipts", "event_id", m.SourceEventID
				}
				query := "UPDATE cafe." + table + " SET outcome=$1 WHERE " + key + "=$2"
				if _, err = owner.Exec(t.Context(), query, broken, identity); err != nil {
					t.Fatal(err)
				}
				called := false
				retry := func() (a.Outcome, error) {
					return store.Execute(t.Context(), m, func(a.Loaded[counter]) (a.Mutation[counter], error) {
						called = true
						return a.Mutation[counter]{}, errors.New("retry must not invoke the decision")
					})
				}
				outcome, failure := retry()
				var rejection *d.Violation
				if failure == nil || errors.As(failure, &rejection) {
					t.Errorf("corrupt receipt must fail transiently: %+v %v", outcome, failure)
				}
				if called {
					t.Error("corrupt receipt caused the decision to run")
				}
				loaded, err := Query[counter](db, "test_counter").Get(t.Context(), id)
				if err != nil || loaded.Version != 1 || loaded.State.Value != 1 {
					t.Errorf("receipt corruption changed the root: %+v %v", loaded, err)
				}
				var count int
				if err = owner.QueryRow(t.Context(), `SELECT count(*) FROM cafe.consumer_receipts WHERE consumer=$1`, m.Consumer).Scan(&count); err != nil {
					t.Fatal(err)
				}
				want := 0
				if receipt == "consumer" {
					want = 1
				}
				if count != want {
					t.Errorf("corrupt outcome was copied to a new receipt: got %d, want %d", count, want)
				}
				if t.Failed() {
					return
				}
				if _, err = owner.Exec(t.Context(), query, original, identity); err != nil {
					t.Fatal(err)
				}
				recovered, err := retry()
				if err != nil || called || !reflect.DeepEqual(recovered, first) {
					t.Fatalf("repair must recover the original outcome without deciding again: %+v %v", recovered, err)
				}
			})
		}
	}
}
