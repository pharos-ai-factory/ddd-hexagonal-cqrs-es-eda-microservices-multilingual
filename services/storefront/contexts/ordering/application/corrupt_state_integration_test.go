//go:build integration

package application_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

func TestStoredOrderPriceMustBeExplicit(t *testing.T) {
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Fatal("corruption fixtures require the disposable integration project")
	}
	db, err := pg.Open(t.Context(), os.Getenv("ORDERING_DATABASE_URL"), "ordering", protobuf.Encode)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err = db.EnableRealtime(t.Context(), realtime.Encode); err != nil {
		t.Fatal(err)
	}
	admin, err := pgxpool.New(t.Context(), os.Getenv("DATABASE_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	for _, name := range []string{"missing", "null", "different_root", "zero", "paid"} {
		t.Run(name, func(t *testing.T) {
			id := pg.NewID()
			selection := map[string]any{"offerCode": "C1", "name": "Coffee"}
			if name != "missing" {
				selection["minor"] = map[string]any{"null": nil, "zero": 0, "paid": 300, "different_root": 300}[name]
			}
			state := map[string]any{"id": id, "customerId": pg.NewID(), "editionId": pg.NewID(),
				"currency": "EUR", "status": "draft", "lines": []any{
					map[string]any{"id": pg.NewID(), "quantity": 1, "selection": selection},
				}}
			if name == "different_root" {
				state["id"] = pg.NewID()
			}
			raw, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := admin.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(t.Context())
			if _, err = tx.Exec(t.Context(), `SELECT set_config('cafe.command_target',$1,true)`, "order:"+id); err != nil {
				t.Fatal(err)
			}
			if _, err = tx.Exec(t.Context(), `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('order',$1,1,$2)`, id, raw); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(t.Context()); err != nil {
				t.Fatal(err)
			}
			m := a.Metadata{ID: pg.NewID(), AggregateID: id, Name: "ordering.PlaceOrder",
				ExpectedVersion: a.Expected(1), CorrelationID: pg.NewID(), Input: struct{}{},
				Consumer: "ordering.test-corrupt-price", SourceEventID: pg.NewID(), SourceHash: "fixture"}
			handler := app.PlaceOrderHandler{Orders: pg.Command[d.State](db, "order")}
			outcome, failure := handler.Execute(t.Context(), m, struct{}{})
			if name == "missing" || name == "null" || name == "different_root" {
				var violation *core.Violation
				if failure == nil || errors.As(failure, &violation) {
					t.Errorf("missing authority must fail transiently, got outcome=%+v error=%v", outcome, failure)
				}
				var version uint64
				var stored []byte
				if err = admin.QueryRow(t.Context(), `SELECT version,state FROM cafe.aggregates WHERE kind='order' AND id=$1`, id).Scan(&version, &stored); err != nil {
					t.Fatal(err)
				}
				var before, after any
				if err = json.Unmarshal(raw, &before); err != nil {
					t.Fatal(err)
				}
				if err = json.Unmarshal(stored, &after); err != nil {
					t.Fatal(err)
				}
				if version != 1 || !reflect.DeepEqual(before, after) {
					t.Error("corrupt authority changed the stored order")
				}
				orderEvidence(t, admin, m, 0)
				if _, err := pg.Query[d.State](db, "order").Get(t.Context(), id); err == nil {
					t.Error("query accepted corrupt authority")
				}
				if t.Failed() {
					return
				}
				// Administrative repair preserves the logical version and retry identity.
				repair, err := admin.Begin(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				defer repair.Rollback(t.Context())
				if _, err = repair.Exec(t.Context(), `SET LOCAL session_replication_role='replica'`); err != nil {
					t.Fatal(err)
				}
				if _, err = repair.Exec(t.Context(), `UPDATE cafe.aggregates
					SET state=jsonb_set(jsonb_set(state,'{lines,0,selection,minor}','300'),'{id}',to_jsonb($1::text))
					WHERE kind='order' AND id=$1::uuid`, id); err != nil {
					t.Fatal(err)
				}
				if err = repair.Commit(t.Context()); err != nil {
					t.Fatal(err)
				}
				outcome, failure = handler.Execute(t.Context(), m, struct{}{})
			}
			if failure != nil || outcome.Rejection != nil || outcome.Status != "placed" || outcome.Version != 2 {
				t.Fatalf("explicit price must permit placement: %+v %v", outcome, failure)
			}
			orderEvidence(t, admin, m, 1)
			loaded, err := pg.Query[d.State](db, "order").Get(t.Context(), id)
			if err != nil {
				t.Fatal(err)
			}
			want := int64(300)
			if name == "zero" {
				want = 0
			}
			if loaded.State.Lines[0].Selection.Minor != want {
				t.Fatalf("placement changed the explicit price: %+v", loaded.State)
			}
		})
	}
}

func orderEvidence(t *testing.T, admin *pgxpool.Pool, m a.Metadata, want int) {
	t.Helper()
	for name, query := range map[string]string{
		"commands":             `SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1`,
		"business events":      `SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=$1`,
		"browser publications": `SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1`,
	} {
		var count int
		if err := admin.QueryRow(t.Context(), query, m.AggregateID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("%s: want %d, got %d", name, want, count)
		}
	}
	var consumers int
	if err := admin.QueryRow(t.Context(), `SELECT count(*) FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2`, m.Consumer, m.SourceEventID).Scan(&consumers); err != nil {
		t.Fatal(err)
	}
	if consumers != want {
		t.Errorf("consumer receipts: want %d, got %d", want, consumers)
	}
}
