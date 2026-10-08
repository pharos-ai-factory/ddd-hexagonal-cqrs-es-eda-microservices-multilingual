// Code generated from Protobuf request contracts. DO NOT EDIT.
export namespace shared {
export interface PageRequest {
  limit: number;
  after?: string;
}
export interface CommandMetadata {
  commandId: string;
  aggregateId: string;
  expectedVersion?: number;
  correlationId: string;
}
export interface Rejection {
  code: string;
  message: string;
}
export interface Outcome {
  aggregateId: string;
  version: number;
  status: string;
  rejection?: Rejection;
}
export interface RequestError {
  code: string;
  message: string;
}
}
export namespace loyalty {
export interface RedeemReward {
  orderId?: string;
}
export interface ListAccounts {
  page?: shared.PageRequest;
}
export interface GetAccount {
  id: string;
}
export interface ListRewards {
  page?: shared.PageRequest;
}
export interface GetReward {
  id: string;
}
export interface Grant {
  id: string;
  accountId: string;
  benefit: string;
  validDays: number;
}
export interface Account {
  id: string;
  stampBalance: number;
  collections: number;
  grantsEarned: number;
  lastGrant?: Grant;
}
export interface Reward {
  id: string;
  grantId: string;
  customerId: string;
  benefit: string;
  status: string;
  expiresAt: string;
  redeemedFor?: string;
}
export interface LoadedAccount {
  exists: boolean;
  version: number;
  state?: Account;
}
export interface Accounts {
  items: LoadedAccount[];
  paged: boolean;
  nextId?: string;
}
export interface LoadedReward {
  exists: boolean;
  version: number;
  state?: Reward;
}
export interface Rewards {
  items: LoadedReward[];
  paged: boolean;
  nextId?: string;
}
}
export namespace communication {
export interface ListNotifications {
  page?: shared.PageRequest;
}
export interface GetNotification {
  id: string;
}
export interface Notification {
  id: string;
  recipient: string;
  subject: string;
  body: string;
  status: string;
  providerReceipt?: string;
}
export interface LoadedNotification {
  exists: boolean;
  version: number;
  state?: Notification;
}
export interface Notifications {
  items: LoadedNotification[];
  paged: boolean;
  nextId?: string;
}
}
