import {readFileSync} from 'node:fs';

/** Resolve NAME or NAME_FILE exclusively; never include values or paths in failures. */
export function secret(name: string): string {
  if (process.env[name] !== undefined && process.env[name + '_FILE'] !== undefined) {
    throw new Error(`${name} and ${name}_FILE are mutually exclusive`);
  }
  let value = process.env[name] ?? '';
  if (process.env[name + '_FILE'] !== undefined) {
    try { value = readFileSync(process.env[name + '_FILE']!, 'utf8').replace(/\r?\n$/, ''); }
    catch { throw new Error(`${name}_FILE cannot be read`); }
  }
  if (!value.trim()) throw new Error(`${name} is required and must not be empty`);
  return value;
}
