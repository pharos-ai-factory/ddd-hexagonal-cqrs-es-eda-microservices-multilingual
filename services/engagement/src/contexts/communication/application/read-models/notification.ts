/** Application-owned read representation of a notification. */
export type NotificationView = Readonly<{
  id: string; recipient: string; subject: string; body: string; status: 'requested' | 'sent'; providerReceipt?: string;
}>;
