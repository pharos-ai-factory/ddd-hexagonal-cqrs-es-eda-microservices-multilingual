// A history gap during an in-flight query requires another query afterwards.
// Coalesce repeated requests, but never discard the need for a fresh snapshot.
export function reconciliationQueue(reconcile: () => Promise<void>) {
  let requested = false, closed = false;
  let running: Promise<void> | undefined;
  return {
    request(): Promise<void> {
      if (closed) return Promise.resolve();
      requested = true;
      if (!running) {
        running = (async () => {
          try {
            while (requested && !closed) {
              requested = false;
              await reconcile();
            }
          } finally { running = undefined; }
        })();
      }
      return running;
    },
    close() { closed = true; requested = false; },
  };
}
