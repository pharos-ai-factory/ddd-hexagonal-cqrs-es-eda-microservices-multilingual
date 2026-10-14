package di

import (
	"context"
	"fmt"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/publications"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/shared"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
	"google.golang.org/protobuf/proto"
	"testing"
)

type observedTransaction[A any] struct {
	transaction execution.Transaction[A]
	active      *bool
	metadata    []a.Metadata
}

func (p *observedTransaction[A]) Execute(ctx context.Context, m a.Metadata, work func(a.WriteRepository[A]) (execution.Result, error)) (a.Outcome, error) {
	p.metadata = append(p.metadata, m)
	*p.active = true
	defer func() { *p.active = false }()
	return p.transaction.Execute(ctx, m, work)
}

type observedDirectory struct {
	t      *testing.T
	active *bool
	calls  int
}

func (p *observedDirectory) Find(_ context.Context, key string) (app.DrinkPublishedDomainEvent, bool, error) {
	p.t.Helper()
	if *p.active {
		p.t.Fatal("projection read acquired a connection inside a transaction")
	}
	p.calls++
	return app.DrinkPublishedDomainEvent{DrinkID: support.Drink, Revision: 1, Name: "Coffee"}, true, nil
}
func (p *observedDirectory) Record(context.Context, a.Metadata, string, uint64, app.DrinkPublishedDomainEvent) error {
	return fmt.Errorf("read only")
}

func TestMenuRegistrationsDispatchPlainHandlersToTheirOwningTransactions(t *testing.T) {
	active := false
	drinks := &support.CommandProbe[d.DrinkState]{}
	editions := &support.CommandProbe[d.EditionState]{}
	drinkTransaction := &observedTransaction[*d.Drink]{transaction: support.ForAggregate(drinks, d.RestoreDrink, (*d.Drink).Snapshot, publications.Drink), active: &active}
	editionTransaction := &observedTransaction[*d.MenuEdition]{transaction: support.ForAggregate(editions, d.RestoreEdition, (*d.MenuEdition).Snapshot, publications.Edition), active: &active}
	directory := &observedDirectory{t: t, active: &active}
	registry := menuRequestsProvider(drinkTransaction, editionTransaction, directory, q.GetDrinkQueryHandler{}, q.ListDrinksQueryHandler{}, q.GetEditionQueryHandler{}, q.ListEditionsQueryHandler{})
	request := &pb.Request{ContractVersion: 1, RequestId: support.Customer, Context: "menu", Payload: &pb.Request_Command{Command: &pb.Command{Metadata: &shared.CommandMetadata{CommandId: support.Customer, AggregateId: support.Drink, ExpectedVersion: proto.Uint64(0), CorrelationId: support.Customer}, Payload: &pb.Command_CreateDrink{CreateDrink: &pb.CreateDrink{Name: proto.String("Coffee")}}}}}
	if reply := registry.Handle(t.Context(), request); reply.GetOutcome() == nil || reply.GetOutcome().Status != "draft" {
		t.Fatalf("create drink: %v", reply)
	}
	if drinks.Loaded.State.Name != "Coffee" || len(drinkTransaction.metadata) != 1 || len(editionTransaction.metadata) != 0 {
		t.Fatal("command reached the wrong aggregate")
	}
	m := drinkTransaction.metadata[0]
	if m.ID != support.Customer || m.ExpectedVersion == nil || *m.ExpectedVersion != 0 || m.Name != "menu.CreateDrink" {
		t.Fatalf("lost receipt metadata: %+v", m)
	}
	request.GetCommand().Payload = &pb.Command_PublishDrink{PublishDrink: &pb.PublishDrink{}}
	if reply := registry.Handle(t.Context(), request); reply.GetOutcome() == nil || reply.GetOutcome().Status != "published" {
		t.Fatalf("publish drink: %v", reply)
	}
	request.GetCommand().Metadata.AggregateId = support.Edition
	request.GetCommand().Payload = &pb.Command_CreateEdition{CreateEdition: &pb.CreateEdition{Currency: proto.String("GBP")}}
	if reply := registry.Handle(t.Context(), request); reply.GetOutcome() == nil || reply.GetOutcome().Status != "draft" {
		t.Fatalf("create edition: %v", reply)
	}
	request.GetCommand().Payload = &pb.Command_AddOffer{AddOffer: &pb.AddOffer{Code: proto.String("COFFEE"), DrinkId: proto.String(support.Drink), DrinkRevision: proto.Uint64(1), Minor: proto.Int64(250)}}
	if reply := registry.Handle(t.Context(), request); reply.GetOutcome() == nil || reply.GetOutcome().GetRejection() != nil {
		t.Fatalf("add offer: %v", reply)
	}
	if directory.calls != 1 || len(editionTransaction.metadata) != 2 || len(drinkTransaction.metadata) != 2 {
		t.Fatal("projection registration bypassed execution policy")
	}
}
