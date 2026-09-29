#!/bin/sh
set -eu
umask 077
. /reference/read-secret.sh
role=${VALKEY_ROLE:-sessions}
case "$role" in
 sessions)
  password=$(read_secret SESSION_PASSWORD)
  acl="user sessions on >${password} ~cafe:auth:* +@read +@write +@connection +@transaction"
  persistence='appendonly yes
appendfsync always'
  ;;
 realtime)
  password=$(read_secret REALTIME_REDIS_PASSWORD)
  acl="user realtime on >${password} ~cafe:realtime* &cafe:realtime* -@all +auth +hello +ping +client|setinfo +eval +evalsha +script|load +publish +subscribe +unsubscribe +del +expire +hget +hmget +hset +hincrby +xadd +xrange +xrevrange"
  # History is disposable. Losing it invokes owner-query reconciliation.
  persistence='appendonly no
save ""'
  ;;
 *) echo 'Unknown Valkey role' >&2; exit 1 ;;
esac
[ -n "$password" ] || { echo 'Valkey credential is required' >&2; exit 1; }
cat > /tmp/cafe-valkey.conf <<CONFIG
bind 0.0.0.0
protected-mode yes
dir /data
$persistence
maxmemory-policy noeviction
user default off
$acl
CONFIG
chown valkey:valkey /tmp/cafe-valkey.conf
exec docker-entrypoint.sh valkey-server /tmp/cafe-valkey.conf
