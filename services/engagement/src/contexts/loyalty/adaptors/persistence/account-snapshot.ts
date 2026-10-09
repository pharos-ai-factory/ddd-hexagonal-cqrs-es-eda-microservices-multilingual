import {storedObject} from '../../../../adaptors/stored-values.js';
import {LoyaltyAccount, type AccountState} from '../../domain/loyalty-account.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreAccountSnapshot = (value: unknown): AccountState => new LoyaltyAccount(storedObject(value) as AccountState).snapshot();
