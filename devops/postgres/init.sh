#!/bin/sh
set -eu
for owner in menu ordering preparation collection loyalty communication; do
 upper=$(printf '%s' "$owner" | tr '[:lower:]' '[:upper:]')
 password=$(printenv "${upper}_DB_PASSWORD")
 psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
  -v role="cafe_${owner}" -v password="$password" -v database="cafe_${owner}" <<'SQL'
CREATE ROLE :"role" LOGIN PASSWORD :'password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE DATABASE :"database";
REVOKE CONNECT ON DATABASE :"database" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"database" TO :"role";
SQL
 psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "cafe_${owner}" -v checksum="$(sha256sum /reference/0001_initial.sql | cut -d ' ' -f 1)" -f /reference/0001_initial.sql
 psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "cafe_${owner}" -v role="cafe_${owner}" <<'SQL'
GRANT USAGE ON SCHEMA cafe TO :"role";
GRANT SELECT ON cafe.schema_version TO :"role";
GRANT SELECT,INSERT,UPDATE ON cafe.aggregates,cafe.dispatches,cafe.projections TO :"role";
GRANT SELECT,INSERT ON cafe.command_receipts,cafe.consumer_receipts,cafe.outbox_events TO :"role";
GRANT EXECUTE ON FUNCTION cafe.guard_aggregate_write() TO :"role";
SQL
 psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "cafe_${owner}" -v role="cafe_${owner}" \
  -v checksum="$(sha256sum /reference/0002_realtime.sql | cut -d ' ' -f 1)" -f /reference/0002_realtime.sql
done
