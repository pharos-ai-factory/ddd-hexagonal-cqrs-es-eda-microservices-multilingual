'use client';
import {useEffect, useState} from 'react';
type Attempt = {path: string; body: object; version: number; key: string; correlation: string};
const storage = 'cafe:pending-command';
export function useCommand(sessionEnded: () => void) {
  const [pending, setPending] = useState<Attempt | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');
  useEffect(() => {
    const saved = sessionStorage.getItem(storage);
    if (saved) { setPending(JSON.parse(saved)); setMessage('A previous command has an uncertain outcome. Retry it to recover the result.'); }
  }, []);
  async function perform(attempt: Attempt) {
    setBusy(true); setPending(attempt);
    sessionStorage.setItem(storage, JSON.stringify(attempt));
    try {
      const response = await fetch('/api/v1/'+attempt.path, {method: 'POST',
        headers: {'Content-Type': 'application/json', 'Idempotency-Key': attempt.key,
          'If-Match': String(attempt.version), 'X-Correlation-ID': attempt.correlation},
        body: JSON.stringify(attempt.body)});
      if (response.status === 401) {
        sessionEnded();
        throw new Error('Sign in again, then retry the same command to recover its outcome.');
      }
      if (response.status >= 500) throw new Error('The outcome is uncertain. Retry the same command.');
      const outcome = await response.json();
      sessionStorage.removeItem(storage); setPending(null);
      if (!response.ok) {
        setMessage(outcome.rejection?.message ?? outcome.message ?? outcome.code ?? 'Command rejected');
        return false;
      }
      setMessage(`Command accepted at version ${outcome.version}. Live projections show the resulting state.`);
      return true;
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'The outcome is uncertain. Retry the same command.');
      return false;
    } finally { setBusy(false); }
  }
  async function send(path: string, body: object, version: number) {
    if (pending || busy) return false;
    let correlation = sessionStorage.getItem('cafe:correlation');
    if (!correlation) { correlation = crypto.randomUUID(); sessionStorage.setItem('cafe:correlation', correlation); }
    return perform({path, body, version, key: crypto.randomUUID(), correlation});
  }
  return {send, busy, pending, message, retry: () => pending && perform(pending)};
}
