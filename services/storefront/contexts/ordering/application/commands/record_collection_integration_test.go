//go:build integration

package commands_test

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	orderpg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/persistence/postgres"
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	reactions "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/eventhandlers"
	ports "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime"
	rt "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime/generated/cafe/realtime/v1"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
	"testing"
)

func TestAcceptedCollectionCommitsOrderReceiptAndExplicitRealtimeTogether(t *testing.T) {
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
	id, customer := pg.NewID(), pg.NewID()
	state := d.OrderState{ID: id, CustomerID: customer, EditionID: pg.NewID(), Currency: "EUR", Status: "placed", Lines: []d.LineState{{ID: pg.NewID(), Selection: d.Selection{OfferCode: "TEA", Name: "Tea", Minor: 200}, Quantity: 1}}}
	raw, _ := json.Marshal(state)
	tx, err := admin.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(t.Context(), `SELECT set_config('cafe.command_target',$1,true)`, "order:"+id); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(t.Context(), `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('order',$1,1,$2)`, id, raw); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	codec := rpc.RecordCollectionCodec{}
	source := pg.NewID()
	m := a.Metadata{ID: a.DerivedID(codec.Consumer(), source), Name: codec.Consumer(), Consumer: codec.Consumer(), SourceEventID: source, CausationID: source, AggregateID: id, CorrelationID: pg.NewID(), SourceHash: strings.Repeat("a", 64), Input: map[string]any{"orderId": id, "customerId": customer}}
	reaction := reactions.OrderCollectedIntegrationEventHandler{Commands: pg.NewCommandOutbox[commands.RecordCollectionCommand](db, codec)}
	if _, err = reaction.Handle(t.Context(), m, model.OrderCollectedIntegrationEvent{OrderID: id, CustomerID: customer}); err != nil {
		t.Fatal(err)
	}
	var body []byte
	if err = admin.QueryRow(t.Context(), `SELECT body FROM cafe.internal_commands WHERE id=$1`, m.ID).Scan(&body); err != nil {
		t.Fatal(err)
	}
	queued, command, err := codec.Decode(body)
	if err != nil {
		t.Fatal(err)
	}
	execute := execution.Bind(orderpg.NewOrderTransaction(db), func(repo ports.OrderWriteRepository, bus a.RealtimeBusPort[a.RealtimeEvent]) commands.RecordCollectionCommandHandler {
		return commands.RecordCollectionCommandHandler{Repository: repo, Realtime: bus}
	})
	// Force failure at browser encoding after the handler has staged its save.
	if err = db.EnableRealtime(t.Context(), func(string, string, string, string, uint64, a.RealtimeEvent) ([]byte, error) {
		return nil, context.DeadlineExceeded
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = execute.Execute(t.Context(), queued, command); err == nil {
		t.Fatal("encoding failure committed")
	}
	loaded, err := orderpg.NewOrderReadRepository(db).Get(t.Context(), id)
	if err != nil || loaded.Version != 1 || loaded.State.CollectionRecorded {
		t.Fatalf("rollback: %+v %v", loaded, err)
	}
	var n int
	if err = admin.QueryRow(t.Context(), `SELECT count(*) FROM cafe.consumer_receipts WHERE event_id=$1`, source).Scan(&n); err != nil || n != 0 {
		t.Fatal("receipt escaped failed command")
	}
	if err = db.EnableRealtime(t.Context(), realtime.Encode); err != nil {
		t.Fatal(err)
	}
	first, err := execute.Execute(t.Context(), queued, command)
	if err != nil || first.Rejection != nil || first.Version != 2 {
		t.Fatalf("execution: %+v %v", first, err)
	}
	second, err := execute.Execute(t.Context(), queued, command)
	if err != nil || first != second {
		t.Fatalf("lost ACK recovery: %+v %v", second, err)
	}
	loaded, err = orderpg.NewOrderReadRepository(db).Get(t.Context(), id)
	if err != nil || !loaded.State.CollectionRecorded || loaded.State.Status != "placed" {
		t.Fatalf("collection not visible: %+v %v", loaded, err)
	}
	if err = admin.QueryRow(t.Context(), `SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1`, id).Scan(&n); err != nil || n != 1 {
		t.Fatal("replay duplicated realtime")
	}
	if err = admin.QueryRow(t.Context(), `SELECT body FROM cafe.realtime_publications WHERE aggregate_id=$1`, id).Scan(&body); err != nil {
		t.Fatal(err)
	}
	var publication rt.Publication
	if err = proto.Unmarshal(body, &publication); err != nil || publication.Revision != 2 || !publication.GetOrder().CollectionRecorded {
		t.Fatalf("browser view: %+v %v", &publication, err)
	}
	// Another source event identity for the same handover produces no second transition.
	queued.SourceEventID = pg.NewID()
	queued.CausationID = queued.SourceEventID
	queued.ID = a.DerivedID(codec.Consumer(), queued.SourceEventID)
	duplicate, err := execute.Execute(t.Context(), queued, command)
	if err != nil || duplicate.Version != 2 {
		t.Fatalf("duplicate business fact: %+v %v", duplicate, err)
	}
	if err = admin.QueryRow(t.Context(), `SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1`, id).Scan(&n); err != nil || n != 1 {
		t.Fatal("duplicate business fact republished realtime")
	}
}
