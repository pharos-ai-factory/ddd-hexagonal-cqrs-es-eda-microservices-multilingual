/** Owns partially composed resources and drains workers before disposing their dependencies. */
export class RuntimeLifecycle {
  readonly controller = new AbortController();
  private resources: Array<() => Promise<void>> = [];
  private tasks: Promise<void>[] = [];
  private stopping?: Promise<void>;
  async acquire<T extends {dispose(): Promise<void>}>(create: () => T | Promise<T>): Promise<T> {
    const value = await create();
    this.resources.push(() => value.dispose());
    return value;
  }
  track(task: Promise<void>): void { this.tasks.push(task); }
  stop(closeListener: () => Promise<void> = async () => {}): Promise<void> {
    return this.stopping ??= (async () => {
      this.controller.abort();
      await Promise.allSettled([...this.tasks, closeListener()]);
      const errors: unknown[] = [];
      for (const dispose of this.resources.reverse()) {
        try { await dispose(); } catch (error) { errors.push(error); }
      }
      if (errors.length) throw new AggregateError(errors, 'Resource disposal failed');
    })();
  }
}
