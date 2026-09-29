//go:build integration

package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestRuntimeRejectsContextSchemaDrift(t *testing.T) {
	// Other packages exercise Ordering concurrently; reserve Menu for schema corruption.
	db, err := Open(t.Context(), os.Getenv("MENU_DATABASE_URL"), "menu", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Database = "cafe_menu"
	administrator, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(administrator.Close)
	for _, change := range []struct{ name, change, restore string }{
		{"identity", "UPDATE cafe.context_identity SET owner='ordering'", "UPDATE cafe.context_identity SET owner='menu'"},
		{"checksum", "UPDATE cafe.context_migrations SET checksum='changed' WHERE version=1", "UPDATE cafe.context_migrations SET checksum=$1 WHERE version=1"},
		{"missing", "DELETE FROM cafe.context_migrations WHERE version=1", "INSERT INTO cafe.context_migrations(version,checksum) VALUES(1,$1)"},
		{"unknown-empty", "UPDATE cafe.context_migrations SET version=999,checksum='' WHERE version=1", "UPDATE cafe.context_migrations SET version=1,checksum=$1 WHERE version=999"},
		{"unexpected", "INSERT INTO cafe.context_migrations(version,checksum) VALUES(999,'unknown')", "DELETE FROM cafe.context_migrations WHERE version=999"},
	} {
		t.Run(change.name, func(t *testing.T) {
			var checksum string
			if err := administrator.QueryRow(t.Context(), `SELECT checksum FROM cafe.context_migrations WHERE version=1`).Scan(&checksum); err != nil {
				t.Fatal(err)
			}
			if _, err := administrator.Exec(t.Context(), change.change); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var args []any
				if change.name == "checksum" || change.name == "missing" || change.name == "unknown-empty" {
					args = append(args, checksum)
				}
				if _, err := administrator.Exec(cleanup, change.restore, args...); err != nil {
					t.Error(err)
				}
			})
			if err := verifyContextSchema(t.Context(), db.pool, "menu"); err == nil {
				t.Fatal("runtime accepted drifted context schema")
			}
		})
	}
	if err := verifyContextSchema(t.Context(), db.pool, "menu"); err != nil {
		t.Fatal(err)
	}
}
