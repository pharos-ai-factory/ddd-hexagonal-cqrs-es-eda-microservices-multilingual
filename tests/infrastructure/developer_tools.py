"""Readiness sees worker outages; inspection follows committed workflow evidence without mutation."""
import os
from pathlib import Path
import subprocess
import sys
import time
sys.path.insert(0, str(Path(__file__).resolve().parents[2]/'scripts'))
from dev import load, compose, wait_ready
from readiness import report
from inspect_workflow import inspect, query


def main():
    env_file = Path(os.environ['CAFE_ENV_FILE'])
    values = load(env_file)
    assert values['COMPOSE_PROJECT_NAME'].startswith('cafe-reference-test-')
    assert report(values)['ready']
    compose(env_file, 'stop', 'operations', stdout=subprocess.DEVNULL)
    try:
        state = report(values, ('OPERATIONS',))
        assert not state['ready']
        deadline = time.monotonic()+15
        while state['checks']['ref.preparation.accept-order.command'] == 'ready':
            assert time.monotonic() < deadline, 'Broker management did not observe consumer removal'
            time.sleep(0.1)
            state = report(values, ('OPERATIONS',))
    finally:
        compose(env_file, 'start', 'operations', stdout=subprocess.DEVNULL)
        wait_ready(values, ('OPERATIONS',))
    rows = query(env_file, 'loyalty', 'SELECT correlation_id FROM cafe.command_receipts ORDER BY created_at LIMIT 1')
    assert rows, 'The completed journey must leave command evidence'
    result = inspect(env_file, rows[0]['correlation_id'])
    assert any(context['commands'] for context in result['contexts'].values())
    assert all('wire' not in row for context in result['contexts'].values() for row in context['commands'])
    assert result['deadQueues']
    print('Development readiness and read-only workflow inspection passed')


if __name__ == '__main__': main()
