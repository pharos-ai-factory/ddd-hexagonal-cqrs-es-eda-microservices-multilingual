'use client';
import {Centrifuge} from 'centrifuge/build/protobuf';
import {useEffect, useReducer, useState} from 'react';
import {decode} from './codec';
import {merge, resources, type Kind, type Projection, type Snapshot} from './model';
import {reconciliationQueue} from './reconciliation';

export function useCafe(active: boolean, sessionEnded: () => void) {
  const [snapshot, dispatch] = useReducer((current: Snapshot, action: Projection | 'clear') =>
    action === 'clear' ? {} : merge(current, action), {});
  const [connection, setConnection] = useState('connecting');
  const [problem, setProblem] = useState('');
  useEffect(() => {
    if (!active) { dispatch('clear'); return; }
    let alive = true, reconciledConnection = false, forceReconcile = false;
    const websocket = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/connection/websocket`;
    const client = new Centrifuge(websocket);
    const reconciliation = reconciliationQueue(async () => {
      try {
        const updates = await Promise.all(Object.entries(resources).map(async ([kind, resource]) => {
          const response = await fetch('/api/v1/'+resource, {cache: 'no-store'});
          if (response.status === 401) { sessionEnded(); throw new Error('Session ended'); }
          if (!response.ok) throw new Error('The current café state could not be loaded. Reconnect to retry.');
          const rows = await response.json() as {version: number; state: Projection['state']}[];
          return rows.map(row => ({...row, kind: kind as Kind}));
        }));
        if (alive) { updates.flat().forEach(update => dispatch(update)); setProblem(''); }
      } catch (error) { if (alive) setProblem(error instanceof Error ? error.message : 'State unavailable'); }
    });
    client.on('connected', () => {
      if (alive) { setConnection('connected'); reconciledConnection = false; }
    });
    client.on('connecting', () => { if (alive) setConnection('connecting'); });
    client.on('subscribed', context => {
      // Each window keeps its own transport cursor. Recoverable history needs no GET.
      if (!reconciledConnection && (forceReconcile || !(context.wasRecovering && context.recovered))) {
        forceReconcile = false;
        reconciledConnection = true;
        void reconciliation.request();
      }
    });
    client.on('publication', context => {
      try { if (alive) dispatch(decode(context.channel, context.data)); }
      catch { if (alive) setProblem('An unrecognised update was rejected. Reconnect to reconcile.'); }
    });
    client.on('disconnected', () => {
      if (!alive) return;
      setConnection('disconnected');
      // Recheck access once after a disconnect, including logout in another window.
      void fetch('/auth/session', {cache: 'no-store'}).then(response => {
        if (alive && response.status === 401) sessionEnded();
      }).catch(() => {});
    });
    const reconnect = () => { forceReconcile = true; client.disconnect(); client.connect(); };
    window.addEventListener('cafe-reconnect', reconnect);
    client.connect();
    return () => {
      alive = false; reconciliation.close();
      window.removeEventListener('cafe-reconnect', reconnect); client.disconnect();
    };
  }, [active, sessionEnded]);
  return {snapshot, connection, problem};
}
