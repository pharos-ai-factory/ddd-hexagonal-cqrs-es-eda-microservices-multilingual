"""Development readiness: database/session access and actual RabbitMQ consumer registration."""
import argparse
import base64
import json
import os
from pathlib import Path
import urllib.request
from urllib.parse import quote
from contract_sources import catalogue, command_subscriptions

OWNERS = {'STOREFRONT': ('menu', 'ordering'), 'OPERATIONS': ('preparation', 'collection'),
          'ENGAGEMENT': ('loyalty', 'communication'), 'API': (), 'WEB': ()}


def request(url, headers=None):
    with urllib.request.urlopen(urllib.request.Request(url, headers=headers or {}), timeout=3) as response:
        data = response.read()
        if url.endswith('/healthz'):
            return {'ready': response.status == 200}
        return json.loads(data)


def broker(values, resource):
    auth = base64.b64encode(('administrator:'+values['BROKER_PASSWORD']).encode()).decode()
    return request(f'http://127.0.0.1:{values["BROKER_HTTP_PORT"]}/api/'+resource, {'Authorization': 'Basic '+auth})


def expected_queues(services, paused=()):
    owners = {owner for service in services for owner in OWNERS[service]}
    queues = {'ref.'+owner+suffix for owner in owners for suffix in ('.commands', '.queries')}
    commands = {item['consumer'] for item in command_subscriptions()}
    for item in catalogue('topology'):
        consumer = item['consumer']
        if consumer.split('.')[0] in owners and consumer not in paused:
            queues.add('ref.'+consumer)
            if consumer in commands:
                queues.add('ref.'+consumer+'.command')
    if 'API' in services:
        queues.update('ref.api.'+owner+'.replies' for service in OWNERS.values() for owner in service)
    return sorted(queues)


def report(values, services=tuple(OWNERS), paused=None, fetch=request, queues_fetch=broker):
    paused = os.environ.get('PAUSED_CONSUMERS', '').split(',') if paused is None else paused
    checks = {}
    for service in services:
        path = '/healthz' if service == 'WEB' else '/diagnostics'
        key = 'API_KEY' if service == 'API' else service+'_API_KEY'
        headers = {} if service == 'WEB' else {'Authorization': 'Bearer '+values[key]}
        try:
            fetch(f'http://127.0.0.1:{values[service+"_PORT"]}'+path, headers)
            checks[service.lower()] = 'ready'
        except Exception:
            checks[service.lower()] = 'database/session/HTTP unavailable'
    if 'WEB' in services:
        try:
            fetch(f'http://127.0.0.1:{values["REALTIME_ADMIN_PORT"]}/health')
            checks['realtime'] = 'ready'
        except Exception:
            checks['realtime'] = 'realtime endpoint unavailable'
    try:
        actual = {item['name']: item for item in queues_fetch(values, 'queues/reference')}
        for queue in expected_queues(services, paused):
            checks[queue] = 'ready' if actual.get(queue, {}).get('consumers', 0) > 0 else 'consumer unavailable'
    except Exception:
        checks['rabbitmq'] = 'management connection unavailable'
    return {'ready': all(value == 'ready' for value in checks.values()), 'checks': checks,
            'pausedSubscriptions': sorted(p for p in paused if p)}


def queue_status(values, queue):
    item = broker(values, 'queues/reference/'+quote(queue, safe=''))
    return {key: item.get(key, 0) for key in ('messages', 'messages_ready', 'messages_unacknowledged', 'consumers')}


def main():
    from dev import ROOT, load
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--env-file', type=Path, default=ROOT/'.local/dev.env')
    args = parser.parse_args()
    try:
        result = report(load(args.env_file))
    except Exception:
        raise SystemExit('Development configuration unavailable; run pnpm dev:doctor and pnpm dev:up') from None
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['ready'] else 1)


if __name__ == '__main__':
    main()
