import { confirmed } from './rabbitmq-publisher.js';
import { failure, errorClass } from '../diagnostics.js';

import amqp from 'amqplib';

import type { PostgresContextDatabase } from '../persistence/postgres-database.js';

import { decode, type WireEvent } from './event-codec.js';

import { claim, finish } from '../persistence/dispatch-leases.js';
import { pause } from '../../foundation/abortable-delay.js';

export async function relay(
  db: PostgresContextDatabase,
  url: string,
  signal: AbortSignal,
  commandHeader?: (body: Buffer) => { id: string; name: string; context: string; correlationId: string },
) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      connection.on('error', (error) => failure(db.owner, 'outbox.connection', error));
      const channel = await connection.createConfirmChannel();
      channel.on('error', (error) => failure(db.owner, 'outbox.channel', error));
      while (!signal.aborted) {
        const row = await claim(db, commandHeader ? 'commands' : false);
        if (!row) {
          await pause(signal);
          continue;
        }
        try {
          const event = commandHeader ? commandHeader(row.body) : decode(row.body);
          await confirmed(
            channel,
            commandHeader ? 'ref.' + db.owner + '.delivery' : 'cafe.events',
            commandHeader ? 'ref.' + event.name : (event as WireEvent).visibility + '.' + event.name,
            row.body,
            {
              contentType: 'application/x-protobuf',
              messageId: event.id,
              type: event.name,
              appId: event.context,
              correlationId: event.correlationId,
              headers: { 'contract-version': { '!': 'int32', value: 1 } },
            },
          );
          await finish(db, row, commandHeader ? 'commands' : false);
        } catch (error) {
          failure(db.owner, 'outbox.publish', error, row.id);
          await finish(db, row, commandHeader ? 'commands' : false, errorClass(error));
          throw error;
        }
      }
    } catch (error) {
      failure(db.owner, 'outbox.reconnect', error);
      await pause(signal, 1000);
    } finally {
      await connection?.close().catch(() => {});
    }
  }
}
