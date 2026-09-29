#!/bin/sh
set -eu
. /reference/read-secret.sh
RABBITMQ_DEFAULT_PASS=$(read_secret BROKER_PASSWORD)
export RABBITMQ_DEFAULT_PASS
exec docker-entrypoint.sh rabbitmq-server
