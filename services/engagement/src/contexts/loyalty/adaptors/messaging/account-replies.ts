import type { Loaded } from '../../../../foundation/application.js';
import type { loyalty as wire } from '../../../../adaptors/generated/request-types.js';
import type { AccountView } from '../../application/read-models/account.js';

export function account(loaded: Loaded<AccountView>): wire.LoadedAccount {
  const s = loaded.state;
  return {
    exists: loaded.exists,
    version: loaded.version,
    state: {
      id: s.id,
      stampBalance: s.stampBalance,
      collections: s.collections,
      grantsEarned: s.grantsEarned,
      ...(s.lastGrant
        ? {
            lastGrant: {
              id: s.lastGrant.id,
              accountId: s.lastGrant.accountId,
              benefit: s.lastGrant.benefit,
              validDays: s.lastGrant.validDays,
            },
          }
        : {}),
    },
  };
}
