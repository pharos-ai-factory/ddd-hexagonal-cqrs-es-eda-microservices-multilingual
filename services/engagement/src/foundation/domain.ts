/** Represents an expected business refusal whose stable code and message can be recorded. */
export class Rejection extends Error {
  constructor(readonly code: string, message: string) { super(message); }
  outcome() { return {code: this.code, message: this.message}; }
}

// Contexts define named subclasses for rules callers need to identify.
/** Base type for named aggregate invariant rejections. */
export abstract class DomainError extends Rejection {}

/** Signals invalid authoritative state so the receiving transaction rolls back. */
export class CorruptState extends Error {
  override name = 'CorruptState';
}

export function identifier(value: string): string {
  if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(value)) {
    throw new Rejection('invalid_id', 'A canonical UUID is required');
  }
  return value;
}
