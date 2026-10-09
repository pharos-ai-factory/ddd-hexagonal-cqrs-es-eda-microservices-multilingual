// A scan starts only after all authorised subscriptions are attached. Publications
// then cover inserts behind a keyset cursor while revision guards cover updates.
export function subscriptionBarrier(channels: readonly string[]) {
  const pending = new Set(channels);
  let needsScan = false;
  return {
    reset(force: boolean) { pending.clear(); channels.forEach(channel => pending.add(channel)); needsScan = force; },
    subscribed(channel: string, recovered: boolean): boolean {
      if (!pending.delete(channel)) return false;
      needsScan ||= !recovered;
      return pending.size === 0 && needsScan;
    },
  };
}
