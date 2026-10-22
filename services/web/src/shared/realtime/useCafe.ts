'use client';
import { useReducer, useState } from 'react';
import { loadCafeState } from './state-loader';
import { mergeState, type VersionedState, type Snapshot } from './model';
import {
  realtimeMessages,
  useRealtimeConnection,
  useRealtimeMessage,
  useRealtimeReconciliation,
} from './index';

/** Café state is an application consumer of the generic typed realtime API. */
export function useCafe() {
  const [snapshot, dispatch] = useReducer(
    (current: Snapshot, update: VersionedState) => mergeState(current, update),
    {},
  );
  const [messageProblem, setMessageProblem] = useState('');
  const connection = useRealtimeConnection();

  useRealtimeMessage(realtimeMessages.drinkChanged, (message) =>
    dispatch({ kind: 'drink', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.editionChanged, (message) =>
    dispatch({ kind: 'edition', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.orderChanged, (message) =>
    dispatch({ kind: 'order', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.ticketChanged, (message) =>
    dispatch({ kind: 'ticket', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.pickupChanged, (message) =>
    dispatch({ kind: 'pickup', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.accountChanged, (message) =>
    dispatch({ kind: 'account', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.rewardChanged, (message) =>
    dispatch({ kind: 'reward', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.notificationChanged, (message) =>
    dispatch({ kind: 'notification', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.operationRejected, (message) =>
    setMessageProblem(`${message.message} (${message.code}; command ${message.commandId})`),
  );
  useRealtimeReconciliation(async (signal) => {
    const updates = await loadCafeState(signal);
    updates.forEach(dispatch);
    setMessageProblem('');
  });

  return {
    snapshot,
    connection: connection.connection,
    problem: connection.problem || messageProblem,
  };
}
