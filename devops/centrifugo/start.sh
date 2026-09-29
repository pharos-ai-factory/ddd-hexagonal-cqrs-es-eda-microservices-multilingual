#!/bin/sh
set -eu
. /reference/read-secret.sh
CENTRIFUGO_HTTP_API_KEY=$(read_secret CENTRIFUGO_API_KEY)
CENTRIFUGO_VAR_CONNECT_PROXY_SECRET=$(read_secret CONNECT_PROXY_SECRET)
redis_password=$(read_secret REALTIME_REDIS_PASSWORD)
CENTRIFUGO_ENGINE_REDIS_ADDRESS="redis://realtime:${redis_password}@realtime-history:6379/0"
export CENTRIFUGO_HTTP_API_KEY CENTRIFUGO_VAR_CONNECT_PROXY_SECRET CENTRIFUGO_ENGINE_REDIS_ADDRESS
exec centrifugo --config=/etc/centrifugo/config.json
