//go:build integration

package commands_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/realtime"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	s "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
)

type projectionWorld struct {
	db          *pg.ContextDatabase
	evidence    *pgxpool.Pool
	menus       *pg.ProjectionStore[model.MenuPublished]
	m           a.Metadata
	input       orderingcommands.CreateOrderCommand
	first, last a.Outcome
	original    a.Loaded[d.OrderState]
}

func (w *projectionWorld) request() error {
	w.m.Input = w.input
	result, err := (orderingcommands.CreateOrderCommandHandler{Orders: pg.Command[d.OrderState](w.db, "order"), Menus: w.menus}).Execute(context.Background(), w.m, w.input)
	w.last = result
	return err
}
func (w *projectionWorld) arrive() error {
	m := a.Metadata{Consumer: "ordering.menu-directory", SourceEventID: pg.NewID(), SourceHash: "scenario-menu"}
	return w.menus.Record(context.Background(), m, w.input.EditionID, 1, model.MenuPublished{EditionID: w.input.EditionID, Currency: "EUR",
		Offers: []model.Offer{{Code: "C1", DrinkID: pg.NewID(), DrinkRevision: 1, Name: "Coffee", Minor: 300, Currency: "EUR"}}})
}
func (w *projectionWorld) noOrder() error {
	state, err := pg.Query[d.OrderState](w.db, "order").Get(context.Background(), w.m.AggregateID)
	if err != nil {
		return err
	}
	if state.Exists {
		return fmt.Errorf("rejection created an order: %+v", state)
	}
	return w.publications(0)
}
func (w *projectionWorld) publications(expected int) error {
	var browser, business int
	err := w.evidence.QueryRow(context.Background(), `SELECT
		(SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1),
		(SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=$1)`, w.m.AggregateID).Scan(&browser, &business)
	if err != nil {
		return err
	}
	if browser != expected || business != 0 {
		return fmt.Errorf("expected %d browser and no business publications; got %d/%d", expected, browser, business)
	}
	return nil
}
func (w *projectionWorld) rejected(code string) error {
	if w.last.Rejection == nil || w.last.Rejection.Code != code {
		return fmt.Errorf("expected %s: %+v", code, w.last)
	}
	return nil
}

func TestProjectionRecoveryFeatures(t *testing.T) {
	url := os.Getenv("ORDERING_DATABASE_URL")
	if url == "" {
		t.Fatal("ORDERING_DATABASE_URL is required; run pnpm test:integration")
	}
	ctx := context.Background()
	db, err := pg.Open(ctx, url, "ordering", protobuf.Encode)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.EnableRealtime(ctx, realtime.Encode); err != nil {
		t.Fatal(err)
	}
	evidence, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer evidence.Close()
	suite := godog.TestSuite{Name: "ordering-postgres",
		Options: s.FeatureOptions(t, "ordering-postgres", "../../../../../../specifications/ordering", "@postgres"),
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			w := &projectionWorld{}
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				zero := uint64(0)
				*w = projectionWorld{db: db, evidence: evidence, menus: pg.Project[model.MenuPublished](db, "menu-directory"),
					m:     a.Metadata{ID: pg.NewID(), AggregateID: pg.NewID(), Name: "create-order", CorrelationID: pg.NewID(), ExpectedVersion: &zero},
					input: orderingcommands.CreateOrderCommand{CustomerID: pg.NewID(), EditionID: pg.NewID()}}
				return ctx, nil
			})
			sc.Step(`^a published menu has not reached Ordering$`, func() error {
				_, found, err := w.menus.Find(ctx, w.input.EditionID)
				if err != nil {
					return err
				}
				if found {
					return fmt.Errorf("scenario menu already exists")
				}
				return nil
			})
			sc.Step(`^a published menu has reached Ordering$`, w.arrive)
			sc.Step(`^the published menu reaches Ordering$`, w.arrive)
			sc.Step(`^the customer requests an order$`, func() error {
				if err := w.request(); err != nil {
					return err
				}
				w.first = w.last
				var err error
				w.original, err = pg.Query[d.OrderState](db, "order").Get(ctx, w.m.AggregateID)
				return err
			})
			sc.Step(`^the decision is recorded as "([^"]*)" without creating an order$`, func(code string) error {
				if err := w.rejected(code); err != nil {
					return err
				}
				var data []byte
				if err := evidence.QueryRow(ctx, `SELECT outcome FROM cafe.command_receipts WHERE aggregate_id=$1 AND command_id=$2`,
					w.m.AggregateID, w.m.ID).Scan(&data); err != nil {
					return err
				}
				var saved a.Outcome
				if err := json.Unmarshal(data, &saved); err != nil {
					return err
				}
				if !reflect.DeepEqual(saved, w.first) {
					return fmt.Errorf("recorded outcome differs: %+v", saved)
				}
				return w.noOrder()
			})
			sc.Step(`^the customer retries the identical command$`, w.request)
			sc.Step(`^the original rejection is returned without creating an order$`, func() error {
				if !reflect.DeepEqual(w.last, w.first) {
					return fmt.Errorf("retry changed the recorded decision: %+v", w.last)
				}
				return w.noOrder()
			})
			sc.Step(`^the customer makes a new attempt$`, func() error { w.m.ID = pg.NewID(); return w.request() })
			sc.Step(`^one draft order and one browser publication are committed$`, func() error {
				if w.last.Rejection != nil || w.last.Version != 1 {
					return fmt.Errorf("new attempt failed: %+v", w.last)
				}
				order, err := pg.Query[d.OrderState](db, "order").Get(ctx, w.m.AggregateID)
				if err != nil {
					return err
				}
				if !order.Exists || order.Version != 1 || order.State.Status != "draft" || order.State.CustomerID != w.input.CustomerID {
					return fmt.Errorf("wrong persisted order: %+v", order)
				}
				return w.publications(1)
			})
			sc.Step(`^the same command identity is reused for another customer$`, func() error {
				if w.first.Rejection != nil {
					return fmt.Errorf("initial creation failed: %+v", w.first)
				}
				w.input.CustomerID = pg.NewID()
				return w.request()
			})
			sc.Step(`^the new input is rejected as "([^"]*)"$`, w.rejected)
			sc.Step(`^the original order and its publication remain unchanged$`, func() error {
				order, err := pg.Query[d.OrderState](db, "order").Get(ctx, w.m.AggregateID)
				if err != nil {
					return err
				}
				if !reflect.DeepEqual(order, w.original) {
					return fmt.Errorf("conflicting input rewrote the order")
				}
				return w.publications(1)
			})
		}}
	if suite.Run() != 0 {
		t.Fatal("Ordering projection recovery scenarios failed")
	}
}
