import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresSnapshotReadRepository} from '../../../../adaptors/snapshot-read-repository.js';
import {MappedReadRepository} from '../../../../adaptors/mapped-read-repository.js';
import {restoreAccountSnapshot} from './account-snapshot.js';
import type {AccountState} from '../../domain/loyalty-account.js';
import type {AccountView} from '../../application/read-models/account.js';

/** Select the stable application read fields independently from storage. */
export const accountView = (s: AccountState): AccountView => ({id: s.id, stampBalance: s.stampBalance, collections: s.collections, grantsEarned: s.grantsEarned,
    ...(s.lastGrant ? {lastGrant: {...s.lastGrant}} : {})});

/** PostgreSQL implementation of the loyalty AccountReadRepository capability. */
export class PostgresAccountReadRepository extends MappedReadRepository<AccountState, AccountView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresSnapshotReadRepository(database, 'account', restoreAccountSnapshot), accountView);
  }
}
