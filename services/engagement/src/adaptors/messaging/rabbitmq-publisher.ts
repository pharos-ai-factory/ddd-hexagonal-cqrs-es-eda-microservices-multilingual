import { type ConfirmChannel, type Options } from 'amqplib';

// One outstanding publication on this channel. Returns precede confirmations.
export async function confirmed(
  channel: ConfirmChannel,
  exchange: string,
  key: string,
  body: Buffer,
  options: Options.Publish,
) {
  let returned = false;
  const onReturn = () => {
    returned = true;
  };
  channel.on('return', onReturn);
  let timeout: ReturnType<typeof setTimeout> | undefined;
  try {
    await Promise.race([
      new Promise<void>((resolve, reject) =>
        channel.publish(
          exchange,
          key,
          body,
          { ...options, persistent: true, mandatory: true },
          (error: unknown) =>
            error
              ? reject(error instanceof Error ? error : new Error('Publisher confirmation failed'))
              : resolve(),
        ),
      ),
      new Promise<void>((_, reject) => {
        timeout = setTimeout(() => reject(new Error('Publisher confirmation timed out')), 5000);
      }),
    ]);
    if (returned) throw new Error('Mandatory publication was returned');
  } finally {
    if (timeout) clearTimeout(timeout);
    channel.removeListener('return', onReturn);
  }
}
