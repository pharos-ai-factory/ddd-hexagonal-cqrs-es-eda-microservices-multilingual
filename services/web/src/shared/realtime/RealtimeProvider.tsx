'use client';

import { Centrifuge } from 'centrifuge/build/protobuf';
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { decodeRealtimeMessage } from '../../adaptors/realtime/message-codec';
import { session } from '../../adaptors/generated/http/session/session';
import { httpStatus, SessionEnded } from '../../adaptors/http/operation-client';
import { owners } from './model';
import type { RealtimeMessage } from './messages';
import { subscriptionBarrier } from './queries';
import { reconciliationQueue } from './reconciliation';

export type RealtimeConnection = 'connecting' | 'connected' | 'disconnected';
type Listener = (message: RealtimeMessage) => void;
type Reconciler = (signal: AbortSignal) => Promise<void>;
type RealtimeContextValue = {
  connection: RealtimeConnection;
  problem: string;
  on: <T extends RealtimeMessage['type']>(
    descriptor: Readonly<{ type: T }>,
    handler: (message: Extract<RealtimeMessage, { type: T }>) => void,
  ) => () => void;
  onReconcile: (handler: Reconciler) => () => void;
};

const RealtimeContext = createContext<RealtimeContextValue | undefined>(undefined);

export function RealtimeProvider({
  principalId,
  sessionEnded,
  children,
}: {
  principalId: string;
  sessionEnded: () => void;
  children: ReactNode;
}) {
  const listeners = useRef(new Map<string, Set<Listener>>());
  const reconcilers = useRef(new Set<Reconciler>());
  const [connection, setConnection] = useState<RealtimeConnection>('connecting');
  const [problem, setProblem] = useState('');
  const on = useCallback<RealtimeContextValue['on']>((descriptor, handler) => {
    const handlers = listeners.current.get(descriptor.type) ?? new Set<Listener>();
    listeners.current.set(descriptor.type, handlers);
    handlers.add(handler as Listener);
    return () => handlers.delete(handler as Listener);
  }, []);
  const onReconcile = useCallback<RealtimeContextValue['onReconcile']>((handler) => {
    reconcilers.current.add(handler);
    return () => reconcilers.current.delete(handler);
  }, []);
  const value = useMemo<RealtimeContextValue>(
    () => ({
      connection,
      problem,
      on,
      onReconcile,
    }),
    [connection, on, onReconcile, problem],
  );

  useEffect(() => {
    let alive = true;
    let forceReconcile = false;
    const seen = new Set<string>();
    const abort = new AbortController();
    const channels = [
      ...new Set(Object.values(owners).map((owner) => `cafe:${owner}`)),
      `principal:${principalId}`,
    ];
    const barrier = subscriptionBarrier(channels);
    const websocket = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/connection/websocket`;
    const client = new Centrifuge(websocket);
    const reconciliation = reconciliationQueue(async () => {
      try {
        await Promise.all([...reconcilers.current].map((handler) => handler(abort.signal)));
        if (alive) setProblem('');
      } catch (error) {
        if (!alive) return;
        if (error instanceof SessionEnded) sessionEnded();
        setProblem(error instanceof Error ? error.message : 'State unavailable');
      }
    });
    client.on('connected', () => {
      if (!alive) return;
      barrier.reset(forceReconcile);
      forceReconcile = false;
      setConnection('connected');
    });
    client.on('connecting', () => {
      if (alive) setConnection('connecting');
    });
    client.on('subscribed', (context) => {
      if (barrier.subscribed(context.channel, context.wasRecovering && context.recovered))
        void reconciliation.request();
    });
    client.on('publication', (context) => {
      try {
        const message = decodeRealtimeMessage(context.channel, context.data);
        if (!alive || seen.has(message.eventId)) return;
        seen.add(message.eventId);
        if (seen.size > 1000) seen.delete(seen.values().next().value!);
        listeners.current.get(message.type)?.forEach((handler) => handler(message));
      } catch {
        if (alive) setProblem('An unrecognised update was rejected. Reconnect to reconcile.');
      }
    });
    client.on('disconnected', () => {
      if (!alive) return;
      setConnection('disconnected');
      void session().catch((error: unknown) => {
        if (alive && httpStatus(error) === 401) sessionEnded();
      });
    });
    const reconnect = () => {
      forceReconcile = true;
      client.disconnect();
      client.connect();
    };
    window.addEventListener('cafe-reconnect', reconnect);
    client.connect();
    return () => {
      alive = false;
      abort.abort();
      reconciliation.close();
      window.removeEventListener('cafe-reconnect', reconnect);
      client.disconnect();
    };
  }, [principalId, sessionEnded]);

  return <RealtimeContext.Provider value={value}>{children}</RealtimeContext.Provider>;
}

function realtimeContext(): RealtimeContextValue {
  const context = useContext(RealtimeContext);
  if (!context) throw new Error('Realtime hooks require RealtimeProvider');
  return context;
}

/** Register a typed handler. Server authorisation determines the subscriptions. */
export function useRealtimeMessage<T extends RealtimeMessage['type']>(
  descriptor: Readonly<{ type: T }>,
  handler: (message: Extract<RealtimeMessage, { type: T }>) => void,
) {
  const { on } = realtimeContext();
  const handlerRef = useRef(handler);
  handlerRef.current = handler;
  useEffect(() => on(descriptor, (message) => handlerRef.current(message)), [descriptor, on]);
}

export function useRealtimeConnection() {
  const { connection, problem } = realtimeContext();
  return { connection, problem };
}

/** Register owner-state recovery performed after all server subscriptions attach. */
export function useRealtimeReconciliation(handler: Reconciler) {
  const { onReconcile } = realtimeContext();
  const handlerRef = useRef(handler);
  handlerRef.current = handler;
  useEffect(() => onReconcile((signal) => handlerRef.current(signal)), [onReconcile]);
}
