#!/bin/sh
set -eu
umask 077
cat > /tmp/cafe-valkey.conf <<CONFIG
bind 0.0.0.0
protected-mode yes
dir /data
appendonly yes
appendfsync always
user default off
user sessions on >${SESSION_PASSWORD} ~cafe:auth:* +@read +@write +@connection +@transaction
# Centrifugo's configured stream history, idempotent publication and Pub/Sub.
# Script execution also checks these command/key/channel permissions.
user realtime on >${REALTIME_REDIS_PASSWORD} ~cafe:realtime* &cafe:realtime* -@all +auth +hello +ping +client|setinfo +eval +evalsha +script|load +publish +subscribe +unsubscribe +del +expire +hget +hmget +hset +hincrby +xadd +xrange +xrevrange
CONFIG
chown valkey:valkey /tmp/cafe-valkey.conf
exec docker-entrypoint.sh valkey-server /tmp/cafe-valkey.conf
