import type { Kind, Models } from './model';

export type RealtimeTarget =
  | Readonly<{ kind: 'context'; id: string }>
  | Readonly<{ kind: 'principal'; id: string }>;

export type StateChangedMessage<K extends Kind = Kind> = {
  [P in K]: {
    type: `${P}Changed`;
    eventId: string;
    target: Readonly<{ kind: 'context'; id: string }>;
    revision: number;
    state: Models[P];
  };
}[K];

export type OperationRejectedMessage = {
  type: 'operationRejected';
  eventId: string;
  target: Readonly<{ kind: 'principal'; id: string }>;
  commandId: string;
  correlationId: string;
  code: string;
  message: string;
};

export type RealtimeMessage = StateChangedMessage | OperationRejectedMessage;

export type RealtimeMessageDescriptor<M extends RealtimeMessage> = Readonly<{
  type: M['type'];
}>;

const descriptor = <M extends RealtimeMessage>(type: M['type']): RealtimeMessageDescriptor<M> =>
  Object.freeze({ type });

/** Typed message catalogue used by browser handlers. It carries no channel choice. */
export const realtimeMessages = Object.freeze({
  drinkChanged: descriptor<StateChangedMessage<'drink'>>('drinkChanged'),
  editionChanged: descriptor<StateChangedMessage<'edition'>>('editionChanged'),
  orderChanged: descriptor<StateChangedMessage<'order'>>('orderChanged'),
  ticketChanged: descriptor<StateChangedMessage<'ticket'>>('ticketChanged'),
  pickupChanged: descriptor<StateChangedMessage<'pickup'>>('pickupChanged'),
  accountChanged: descriptor<StateChangedMessage<'account'>>('accountChanged'),
  rewardChanged: descriptor<StateChangedMessage<'reward'>>('rewardChanged'),
  notificationChanged: descriptor<StateChangedMessage<'notification'>>('notificationChanged'),
  operationRejected: descriptor<OperationRejectedMessage>('operationRejected'),
});
