package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

//go:embed migrations/0001_initial.sql
var schemaSQL []byte

//go:embed migrations/0002_realtime.sql
var realtimeSchemaSQL []byte

type Database struct {
	pool     *pgxpool.Pool
	owner    string
	encode   a.Encoder
	realtime a.RealtimeEncoder
}

func Open(ctx context.Context, url, owner string, encode a.Encoder) (*Database, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	var actual string
	var privileged bool
	var version int
	var checksum string
	err = pool.QueryRow(ctx, `SELECT current_database(),rolsuper OR rolcreatedb OR rolcreaterole OR has_database_privilege(current_user,current_database(),'CREATE') OR has_schema_privilege(current_user,'cafe','CREATE') FROM pg_roles WHERE rolname=current_user`).Scan(&actual, &privileged)
	if err != nil || privileged || actual != "cafe_"+owner {
		pool.Close()
		return nil, fmt.Errorf("database ownership or runtime privilege check failed for %s", owner)
	}
	if err = pool.QueryRow(ctx, `SELECT version, checksum FROM cafe.schema_version`).Scan(&version, &checksum); err != nil || version != 1 || checksum != fmt.Sprintf("%x", sha256.Sum256(schemaSQL)) {
		pool.Close()
		return nil, fmt.Errorf("database schema is not at version 1")
	}
	if err = verifyContextSchema(ctx, pool, owner); err != nil {
		pool.Close()
		return nil, err
	}
	return &Database{pool: pool, owner: owner, encode: encode}, nil
}
func (db *Database) Close() { db.pool.Close() }
func (db *Database) EnableRealtime(ctx context.Context, encoder a.RealtimeEncoder) error {
	var checksum string
	if err := db.pool.QueryRow(ctx, `SELECT checksum FROM cafe.schema_migrations WHERE version=2`).Scan(&checksum); err != nil || checksum != fmt.Sprintf("%x", sha256.Sum256(realtimeSchemaSQL)) {
		return fmt.Errorf("realtime schema is not at version 2")
	}
	if _, err := db.pool.Exec(ctx, `SELECT id FROM cafe.realtime_publications LIMIT 0`); err != nil {
		return err
	}
	db.realtime = encoder
	return nil
}
func NewID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	raw[6] = (raw[6] & 15) | 64
	raw[8] = (raw[8] & 63) | 128
	return formatID(raw[:])
}

// DerivedID gives an idempotent platform operation a stable identity. It never
// derives identities from display names, provider identifiers or mutable data.
// SHA-256 custom derivation uses UUID version 8, rather than claiming UUIDv5.
func DerivedID(purpose, key string) string {
	sum := sha256.Sum256([]byte("cafe-reference/v1\x00" + purpose + "\x00" + key))
	sum[6] = (sum[6] & 15) | 128
	sum[8] = (sum[8] & 63) | 128
	return formatID(sum[:16])
}
func formatID(raw []byte) string {
	h := hex.EncodeToString(raw)
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}
func fingerprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
func begin(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, error) {
	return pool.BeginTx(ctx, pgx.TxOptions{})
}
