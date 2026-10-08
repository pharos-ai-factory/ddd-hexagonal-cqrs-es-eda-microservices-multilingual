'use client';
import {useEffect, useState} from 'react';
import {commandRequest, success} from '../../adaptors/http/client';
import {routes, type CommandOperation, type CommandBody, type SendCommand} from '../../adaptors/generated/http';
type Attempt = {operation: CommandOperation; id: string; body: CommandBody<CommandOperation>; version: number; key: string; correlation: string};
const storage = 'cafe:pending-command';
export function useCommand(sessionEnded: () => void) {
  const [pending, setPending] = useState<Attempt | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  useEffect(() => {
    const saved = sessionStorage.getItem(storage);
    if (saved) { setPending(restoreAttempt(JSON.parse(saved))); setMessage('A previous command has an uncertain outcome. Retry it to recover the result.'); }
  }, []);
  async function perform(attempt: Attempt) {
    setBusy(true); setPending(attempt);
    sessionStorage.setItem(storage, JSON.stringify(attempt));
    try {
      const response = await commandRequest(attempt.operation, attempt.body, {path: {id: attempt.id}, headers: {'Idempotency-Key': attempt.key, 'If-Match': String(attempt.version),
          'X-Correlation-ID': attempt.correlation}});
      if (response.status === 401) {
        sessionEnded();
        throw new Error('Sign in again, then retry the same command to recover its outcome.');
      }
      if (response.status >= 500) throw new Error('The outcome is uncertain. Retry the same command.');
      if (!response.ok) {
        const rejected = await response.json();
        sessionStorage.removeItem(storage); setPending(null);
        setMessage(rejected.rejection?.message ?? rejected.message ?? rejected.code ?? 'Command rejected');
        return false;
      }
      const outcome = await success(attempt.operation, response);
      sessionStorage.removeItem(storage); setPending(null);
      setMessage(`Command accepted at version ${outcome.version}. Live projections show the resulting state.`);
      return true;
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'The outcome is uncertain. Retry the same command.');
      return false;
    } finally { setBusy(false); }
  }
  const send: SendCommand = async (operation, id, body, version) => {
    if (pending || busy) return false;
    let correlation = sessionStorage.getItem('cafe:correlation');
    if (!correlation) { correlation = crypto.randomUUID(); sessionStorage.setItem('cafe:correlation', correlation); }
    return perform({operation, id, body, version, key: crypto.randomUUID(), correlation});
  }
  return {send, busy, pending, message, retry: () => pending && perform(pending)};
}

// Retain attempts saved by the previous frontend until their outcome is recovered.
function restoreAttempt(value: Attempt & {path?: string}): Attempt {
  if (value.operation) return value;
  for (const [operation, route] of Object.entries(routes)) {
    if (route.method !== 'POST' || !route.path.startsWith('/api/v1/')) continue;
    const [prefix, suffix] = route.path.slice('/api/v1/'.length).split('{id}');
    if (suffix === undefined || !value.path?.startsWith(prefix) || !value.path.endsWith(suffix)) continue;
    const id = value.path.slice(prefix.length, suffix ? -suffix.length : undefined);
    if (!id.includes('/')) return {...value, operation: operation as CommandOperation, id};
  }
  throw new Error('The saved command operation is unavailable');
}
