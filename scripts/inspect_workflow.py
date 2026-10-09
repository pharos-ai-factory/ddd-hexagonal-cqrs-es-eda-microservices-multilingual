"""Read development workflow evidence by command, event or correlation UUID without consuming messages."""
import argparse
import json
from pathlib import Path
import subprocess
from uuid import UUID
from dev import ROOT, OWNERS, compose, load
from readiness import queue_status
from contract_sources import catalogue, command_subscriptions


def envelope_identity(body):
    """Read only stable string header fields; never display receipt material or command payloads."""
    offset, fields = 0, {}
    def varint():
        nonlocal offset
        value = 0
        for shift in range(0, 70, 7):
            if offset >= len(body):
                raise ValueError('Truncated envelope')
            byte = body[offset]
            offset += 1
            value |= (byte & 127) << shift
            if byte < 128:
                return value
        raise ValueError('Invalid varint')
    while offset < len(body):
        tag = varint()
        number, kind = tag >> 3, tag & 7
        if kind == 2:
            length = varint()
            end = offset+length
            if end > len(body):
                raise ValueError('Truncated field')
            if number in (1, 2, 3, 5, 6):
                fields[number] = body[offset:end].decode('utf8')
            offset = end
        elif kind == 0:
            varint()
        elif kind in (1, 5):
            offset += 8 if kind == 1 else 4
            if offset > len(body):
                raise ValueError('Truncated fixed field')
        else:
            raise ValueError('Unsupported envelope field')
    return {'commandId': str(UUID(fields[1])), 'consumer': fields[2], 'eventId': str(UUID(fields[3])),
            'target': str(UUID(fields[5])), 'correlationId': str(UUID(fields[6]))}


def query(env_file, owner, statement):
    sql = "BEGIN READ ONLY; SET LOCAL statement_timeout='3s'; SELECT coalesce(json_agg(row_to_json(q)), '[]') FROM ("+statement+") q; COMMIT;"
    result = compose(env_file, 'exec', '-T', 'postgres', 'psql', '-X', '-qAt', '-U', 'postgres', '-d', 'cafe_'+owner,
                     '-v', 'ON_ERROR_STOP=1', '-c', sql, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    return json.loads(result.stdout)


def inspect(env_file, identity):
    ids = {str(UUID(identity))}
    values, owners = load(env_file), {}
    for _ in range(2):
        quoted = ','.join("'"+str(UUID(value))+"'" for value in sorted(ids))
        for owner in OWNERS:
            receipts = query(env_file, owner, f"""SELECT command_id,command_name,aggregate_id,correlation_id,causation_id,
                outcome,created_at FROM cafe.command_receipts WHERE command_id IN ({quoted})
                OR correlation_id IN ({quoted}) OR causation_id IN ({quoted}) ORDER BY created_at LIMIT 201""")
            events = query(env_file, owner, f"""SELECT o.id,o.event_name,o.correlation_id,o.causation_id,d.completed_at AS published_at,
                d.last_error FROM cafe.outbox_events o JOIN cafe.dispatches d ON d.event_id=o.id
                WHERE o.id IN ({quoted}) OR o.correlation_id IN ({quoted}) OR o.causation_id IN ({quoted}) ORDER BY o.created_at LIMIT 201""")
            commands = []
            if owner in ('preparation', 'collection', 'loyalty', 'communication'):
                matches = ' OR '.join("position(convert_to('"+str(UUID(value))+"','UTF8') in c.body)>0" for value in sorted(ids))
                commands = query(env_file, owner, f"""SELECT encode(c.body,'hex') AS wire,c.created_at AS accepted_at,
                    d.completed_at AS published_at,d.last_error,r.outcome,r.completed_at AS executed_at
                    FROM cafe.internal_commands c JOIN cafe.internal_command_dispatches d ON d.event_id=c.id
                    LEFT JOIN cafe.consumer_receipts r ON r.consumer=c.consumer AND r.event_id=c.event_id
                    WHERE {matches} ORDER BY c.created_at LIMIT 201""")
                for command in commands:
                    command.update(envelope_identity(bytes.fromhex(command.pop('wire'))))
                    command['state'] = ('rejected' if command['outcome'].get('rejection') else 'completed') if command['executed_at'] else (
                        'published; execution outcome pending' if command['published_at'] else 'accepted; publication pending')
                    ids.add(command['correlationId'])
            for row in receipts+events:
                ids.add(row['correlation_id'])
            owners[owner] = {'receipts': receipts[:200], 'events': events[:200], 'commands': commands[:200],
                            'truncated': any(len(rows)>200 for rows in (receipts, events, commands))}
    queues = {}
    commands = {item['consumer'] for item in command_subscriptions()}
    for item in catalogue('topology'):
        consumer = item['consumer']
        for suffix in ('', '.command') if consumer in commands else ('',):
            name = 'ref.'+consumer+suffix+'.dead'
            queues[name] = queue_status(values, name)
    return {'identity': identity, 'contexts': owners, 'deadQueues': queues,
            'deadQueueScope': 'Queue-wide counts; a pending command is not proof that this message is dead-lettered.'}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('identity', type=lambda value: str(UUID(value)))
    parser.add_argument('--env-file', type=Path, default=ROOT/'.local/dev.env')
    args = parser.parse_args()
    try:
        print(json.dumps(inspect(args.env_file, args.identity), indent=2))
    except Exception as error:
        raise SystemExit('Workflow inspection unavailable ('+type(error).__name__+'); run pnpm dev:status') from None


if __name__ == '__main__':
    main()
