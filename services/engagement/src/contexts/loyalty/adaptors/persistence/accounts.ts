import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresAggregateQueries} from '../../../../adaptors/postgres.js';
import {MappedQueries} from '../../../../adaptors/mapped-queries.js';
import {storedObject} from '../../../../adaptors/stored-values.js';
import {LoyaltyAccount, type AccountState} from '../../domain/loyalty-account.js';
import type {AccountView} from '../../application/read-models/account.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreAccount = (value: unknown): AccountState => new LoyaltyAccount(storedObject(value) as AccountState).snapshot();
/** Select the stable application read fields independently from storage. */
export const accountView = (s: AccountState): AccountView => ({id: s.id, stampBalance: s.stampBalance, collections: s.collections, grantsEarned: s.grantsEarned,
    ...(s.lastGrant ? {lastGrant: {...s.lastGrant}} : {})});

/** PostgreSQL implementation of the loyalty AccountReader capability. */
export class PostgresAccountReader extends MappedQueries<AccountState, AccountView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresAggregateQueries(database, 'account', restoreAccount), accountView);
  }
}
