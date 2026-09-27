import type {Outcome} from '../foundation/application.js';
import {CorruptState} from '../foundation/domain.js';

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new CorruptState('Corrupt recorded command outcome');
  }
  return value as Record<string, unknown>;
}

export function outcome(value: unknown, target: string): Outcome {
  const fields = record(value);
  if (fields.aggregateId !== target || typeof fields.version !== 'number' ||
      !Number.isSafeInteger(fields.version) || fields.version < 0 || typeof fields.status !== 'string') {
    throw new CorruptState('Corrupt recorded command outcome');
  }
  const result: Outcome = {aggregateId: target, version: fields.version, status: fields.status};
  if ('rejection' in fields) {
    const rejection = record(fields.rejection);
    if (typeof rejection.code !== 'string' || !rejection.code ||
        typeof rejection.message !== 'string' || !rejection.message) {
      throw new CorruptState('Corrupt recorded rejection');
    }
    result.rejection = {code: rejection.code, message: rejection.message};
  } else if (!result.status) {
    throw new CorruptState('Missing recorded outcome status');
  }
  return result;
}

export function checkRootIdentity(state: unknown, target: string): void {
  if (record(state).id !== target) throw new CorruptState('Snapshot identity differs from its storage key');
}
