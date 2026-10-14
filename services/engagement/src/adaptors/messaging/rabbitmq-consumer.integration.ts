import { test } from 'node:test';
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { setTimeout } from 'node:timers/promises';
import amqp from 'amqplib';
import { consume } from './rabbitmq-consumer.js';
import { confirmed } from './rabbitmq-publisher.js';
import { encode } from './event-codec.js';

for (const { counter, malformed } of [
  { counter: 2147483647, malformed: true },
  ...[2147483647, -2147483648, 'invalid', 1.5, true, null].map((counter) => ({ counter, malformed: false })),
]) {
  void test(`invalid retry counter ${String(counter)} with ${malformed ? 'malformed' : 'valid'} bytes reaches quarantine`, async () => {
    const connection = await amqp.connect(process.env.BROKER_ADMIN_URL!);
    const channel = await connection.createConfirmChannel();
    // Runtime workers are stopped; retain the generated queue topology and owner credential.
    const consumer = 'loyalty.issue-reward',
      queue = 'ref.' + consumer;
    const controller = new AbortController();
    let worker: Promise<void> | undefined;
    const id = randomUUID();
    const { event, body: encoded } = encode(
      'loyalty',
      'account',
      id,
      1,
      { id: randomUUID(), target: id, name: 'test', correlation: id, input: {} },
      {
        name: 'loyalty.reward-earned',
        payload: { accountId: id, grantId: randomUUID(), benefit: 'coffee', validDays: 7 },
      },
    );
    const body = malformed ? Buffer.from([0]) : encoded;
    let routed = 0,
      handled = 0;
    try {
      await confirmed(channel, 'ref.loyalty.delivery', queue, body, {
        messageId: event.id,
        contentType: 'application/x-protobuf',
        type: event.name,
        appId: event.context,
        correlationId: event.correlationId,
        headers: {
          'contract-version': 1,
          'ref-attempt':
            typeof counter === 'number' && Number.isInteger(counter)
              ? { '!': 'int32', value: counter }
              : counter,
        },
      });
      worker = consume(
        process.env.LOYALTY_BROKER_URL!,
        {
          owner: 'loyalty',
          consumer,
          event: 'loyalty.reward-earned',
          target: () => {
            routed++;
            return id;
          },
          handle: async () => {
            handled++;
            return { aggregateId: id, version: 1, status: 'issued' };
          },
        },
        controller.signal,
      );
      const deadline = Date.now() + 5000;
      let quarantined = false;
      while (Date.now() < deadline) {
        const message = await channel.get(queue + '.dead', { noAck: false });
        if (message) {
          assert.equal(message.properties.messageId, event.id);
          assert.deepEqual(message.content, body);
          assert.ok(Number.isInteger(message.properties.headers?.['ref-attempt']));
          channel.ack(message);
          quarantined = true;
          break;
        }
        await setTimeout(50);
      }
      assert.ok(quarantined, 'Retry metadata overflow caused redelivery instead of quarantine');
      assert.equal(routed, 0, 'Invalid retry metadata reached routing');
      assert.equal(handled, 0, 'Invalid retry metadata reached a handler');
    } finally {
      controller.abort();
      await worker;
      for (const suffix of ['', '.retry', '.dead']) await channel.purgeQueue(queue + suffix);
      await connection.close();
    }
  });
}
