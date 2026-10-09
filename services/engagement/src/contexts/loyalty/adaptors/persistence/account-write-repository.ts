import type {WriteRepository} from '../../../../foundation/write-repository.js';

import {MappedWriteRepository} from '../../../../adaptors/mapped-write-repository.js';
import {LoyaltyAccount, type AccountState} from '../../domain/loyalty-account.js';

/** Restore authoritative aggregates and stage snapshots in the active PostgreSQL transaction. */
export class PostgresAccountWriteRepository extends MappedWriteRepository<AccountState, LoyaltyAccount> {
  constructor(source: WriteRepository<AccountState>) {
    super(source, state => new LoyaltyAccount(state), aggregate => aggregate.snapshot());
  }
}
