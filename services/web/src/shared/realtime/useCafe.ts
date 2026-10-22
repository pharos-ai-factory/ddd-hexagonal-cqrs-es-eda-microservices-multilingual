'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useMemo, useState } from 'react';
import { loadCafeState } from './state-loader';
import { mergeState, type VersionedState, type Snapshot } from './model';
import {
  realtimeMessages,
  useRealtimeConnection,
  useRealtimeMessage,
  useRealtimeReconciliation,
} from './index';

/** Café state is an application consumer of generated HTTP queries and typed realtime messages. */
export function useCafe(principalId: string) {
  const queryClient = useQueryClient();
  const queryKey = useMemo(() => ['cafe-state', principalId] as const, [principalId]);
  const { data: snapshot } = useQuery<Snapshot>({
    queryKey,
    queryFn: () => Promise.resolve({}),
    enabled: false,
    initialData: {},
    staleTime: Number.POSITIVE_INFINITY,
  });
  const [messageProblem, setMessageProblem] = useState('');
  const connection = useRealtimeConnection();
  const apply = useCallback(
    (update: VersionedState) =>
      queryClient.setQueryData<Snapshot>(queryKey, (current) => mergeState(current ?? {}, update)),
    [queryClient, queryKey],
  );

  useRealtimeMessage(realtimeMessages.drinkChanged, (message) =>
    apply({ kind: 'drink', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.editionChanged, (message) =>
    apply({ kind: 'edition', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.orderChanged, (message) =>
    apply({ kind: 'order', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.ticketChanged, (message) =>
    apply({ kind: 'ticket', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.pickupChanged, (message) =>
    apply({ kind: 'pickup', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.accountChanged, (message) =>
    apply({ kind: 'account', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.rewardChanged, (message) =>
    apply({ kind: 'reward', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.notificationChanged, (message) =>
    apply({ kind: 'notification', version: message.revision, state: message.state }),
  );
  useRealtimeMessage(realtimeMessages.operationRejected, (message) =>
    setMessageProblem(`${message.message} (${message.code}; command ${message.commandId})`),
  );
  useRealtimeReconciliation(async (signal) => {
    const updates = await loadCafeState(queryClient, signal);
    queryClient.setQueryData<Snapshot>(queryKey, (current) => updates.reduce(mergeState, current ?? {}));
    setMessageProblem('');
  });

  return {
    snapshot,
    connection: connection.connection,
    problem: connection.problem || messageProblem,
  };
}
