import type {WriteRepository} from '../../../../foundation/write-repository.js';

import type {LoyaltyAccount} from '../../domain/loyalty-account.js';

/** Authoritative account aggregate persistence inside its command transaction. */
export type AccountWriteRepository = WriteRepository<LoyaltyAccount>;
