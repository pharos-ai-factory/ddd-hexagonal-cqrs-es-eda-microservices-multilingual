import {createHash, randomUUID} from 'node:crypto';
export const newId = () => randomUUID();
export function derivedId(purpose: string, key: string): string {
  const bytes = createHash('sha256').update(`cafe-reference/v1\0${purpose}\0${key}`).digest().subarray(0, 16);
  bytes[6] = (bytes[6]! & 15) | 128;
  bytes[8] = (bytes[8]! & 63) | 128;
  const hex = bytes.toString('hex');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}
