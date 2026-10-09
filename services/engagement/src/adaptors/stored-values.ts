import {CorruptState} from '../foundation/domain.js';

/** Check the storage representation before an owner restores its aggregate. */
export function storedObject(value: unknown): object {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new CorruptState('Stored aggregate snapshot is not an object');
  return value;
}
