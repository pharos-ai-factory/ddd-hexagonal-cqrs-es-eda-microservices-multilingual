import { claimReply, finishReply } from '../persistence/command-reply-outbox.js';

import amqp from 'amqplib';

import type { PostgresContextDatabase } from '../persistence/postgres-database.js';

import { replyRoute } from './reply-address.js';

import { confirmed } from './rabbitmq-publisher.js';

import { failure } from '../diagnostics.js';

import { pause } from '../../foundation/abortable-delay.js';

export async function replyRelay(db: PostgresContextDatabase, url: string, signal: AbortSignal) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      connection.on('error', (error) => failure(db.owner, 'replies.connection', error));
      const channel = await connection.createConfirmChannel();
      channel.on('error', (error) => failure(db.owner, 'replies.channel', error));
      while (!signal.aborted) {
        const row = await claimReply(db);
        if (!row) {
          await pause(signal);
          continue;
        }
        let reason: string | null = null;
        try {
          if (row.expired) reason = 'expired';
          else
            await confirmed(channel, 'cafe.replies', replyRoute(db.owner, row.reply_route), row.body, {
              contentType: 'application/x-protobuf',
              type: 'reply',
              appId: db.owner,
              messageId: row.id,
              correlationId: row.id,
            });
        } catch (error) {
          reason = 'publication_failed';
          throw error;
        } finally {
          await finishReply(db, row, reason);
        }
      }
    } catch (error) {
      failure(db.owner, 'replies.reconnect', error);
      await pause(signal, 1000);
    } finally {
      await connection?.close().catch(() => {});
    }
  }
}
