package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Generated technical metadata contains only this service's context checksums.
//
//go:embed migrations/contexts.json
var contextSchemas []byte

func verifyContextSchema(ctx context.Context, pool *pgxpool.Pool, owner string) error {
	var schemas map[string]map[int]string
	if err := json.Unmarshal(contextSchemas, &schemas); err != nil {
		return err
	}
	expected, known := schemas[owner]
	if !known {
		return fmt.Errorf("unknown schema owner %s", owner)
	}
	var actual string
	if err := pool.QueryRow(ctx, `SELECT owner FROM cafe.context_identity WHERE singleton`).Scan(&actual); err != nil || actual != owner {
		return fmt.Errorf("context schema identity mismatch for %s", owner)
	}
	rows, err := pool.Query(ctx, `SELECT version,checksum FROM cafe.context_migrations`)
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var version int
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return err
		}
		expectedChecksum, known := expected[version]
		if !known || expectedChecksum != checksum {
			return fmt.Errorf("context migration checksum mismatch for %s version %d", owner, version)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if count != len(expected) {
		return fmt.Errorf("context migrations are incomplete for %s", owner)
	}
	return nil
}
