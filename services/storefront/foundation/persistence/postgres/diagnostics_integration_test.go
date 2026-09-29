//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"net/url"
	"os"
	"testing"
)

func TestDiagnosticsReportsDurableBacklogAndUnavailableAuthority(t *testing.T) {
	// Menu has no concurrent application-package fixture writers. Keep exact
	// before/after evidence independent of Ordering's live receipt tests.
	raw := os.Getenv("MENU_DATABASE_URL")
	if raw == "" {
		t.Fatal("integration lane requires MENU_DATABASE_URL")
	}
	db, err := Open(t.Context(), raw, "menu", func(m a.Message) ([]byte, error) { return json.Marshal(m) })
	if err != nil {
		t.Fatal("Menu fixture database unavailable")
	}
	t.Cleanup(db.Close)
	adminURL, err := url.Parse(os.Getenv("DATABASE_ADMIN_URL"))
	if err != nil || adminURL.Host == "" {
		t.Fatal("integration lane requires DATABASE_ADMIN_URL")
	}
	adminURL.Path = "/cafe_menu"
	owner, err := pgxpool.New(t.Context(), adminURL.String())
	if err != nil {
		t.Fatal("Menu fixture administrator unavailable")
	}
	t.Cleanup(owner.Close)
	id := NewID()
	clean(t, owner, id)
	if err := db.EnableRealtime(t.Context(), func(_, _, _, _ string, _ uint64, state []byte) ([]byte, error) { return state, nil }); err != nil {
		t.Fatal(err)
	}
	before, err := db.Diagnostics(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Command[counter](db, "test_counter").Execute(t.Context(), metadata(id, 0), func(a.Loaded[counter]) (a.Mutation[counter], error) {
		return a.Changed(counter{1}, "active", a.Publication{Name: "test.created", Visibility: a.Private, Payload: counter{1}}), nil
	}); err != nil {
		t.Fatal(err)
	}
	after, err := db.Diagnostics(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if after.Outbox.Pending != before.Outbox.Pending+1 || after.Realtime.Pending != before.Realtime.Pending+1 {
		t.Fatalf("missing durable intent: before=%+v after=%+v", before, after)
	}
	if after.Outbox.OldestAgeSeconds < 0 || after.Realtime.OldestAgeSeconds < 0 {
		t.Fatal("invalid backlog ages")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := db.Diagnostics(ctx); err == nil {
		t.Fatal("unavailable authority reported as a zero backlog")
	}
}
