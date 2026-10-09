'use client';
import {Centrifuge} from 'centrifuge/build/protobuf';
import {useEffect, useReducer, useState} from 'react';
import {decode} from './codec';
import {merge, owners, type Projection, type Snapshot} from './model';
import {subscriptionBarrier} from './queries';
import {SessionEnded} from '../../adaptors/http/client';
import {projections} from './httpProjection';
import {request} from '../../adaptors/http/client';
import {reconciliationQueue} from './reconciliation';

export function useCafe(active: boolean, sessionEnded: () => void) {
  const [snapshot, dispatch] = useReducer((current: Snapshot, action: Projection | 'clear') =>
    action === 'clear' ? {} : merge(current, action), {});
  const [connection, setConnection] = useState('connecting');
  const [problem, setProblem] = useState('');
  useEffect(() => {
    if (!active) { dispatch('clear'); return; }
    let alive = true, forceReconcile = false;
    const abort = new AbortController();
    const barrier = subscriptionBarrier([...new Set(Object.values(owners).map(owner => `cafe:${owner}`))]);
    const websocket = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/connection/websocket`;
    const client = new Centrifuge(websocket);
    const reconciliation = reconciliationQueue(async () => {
      try {
        const updates = await projections(abort.signal);
        if (alive) { updates.forEach(update => dispatch(update)); setProblem(''); }
      } catch (error) { if (alive) {
        if (error instanceof SessionEnded) sessionEnded();
        setProblem(error instanceof Error ? error.message : 'State unavailable');
      } }
    });
    client.on('connected', () => {
      if (alive) { barrier.reset(forceReconcile); forceReconcile = false; setConnection('connected'); }
    });
    client.on('connecting', () => { if (alive) setConnection('connecting'); });
    client.on('subscribed', context => {
      // Each window keeps its own transport cursor. Recoverable history needs no GET.
      if (barrier.subscribed(context.channel, context.wasRecovering && context.recovered)) {
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
      void request('session', {}).then(response => {
        if (alive && response.status === 401) sessionEnded();
      }).catch(() => {});
    });
    const reconnect = () => { forceReconcile = true; client.disconnect(); client.connect(); };
    window.addEventListener('cafe-reconnect', reconnect);
    client.connect();
    return () => {
      alive = false; abort.abort(); reconciliation.close();
      window.removeEventListener('cafe-reconnect', reconnect); client.disconnect();
    };
  }, [active, sessionEnded]);
  return {snapshot, connection, problem};
}
