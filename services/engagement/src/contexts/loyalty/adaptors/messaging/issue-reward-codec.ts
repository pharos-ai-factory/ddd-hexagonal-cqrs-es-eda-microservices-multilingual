import { LoyaltySubscription } from './loyalty-subscriptions.js';
import { fields } from './command-fields.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { IssueRewardCommand } from '../../application/commands/issue-reward.js';

export const issueCodec = () =>
  new InternalCommandCodec(
    'loyalty',
    LoyaltySubscription.IssueReward,
    'issueReward',
    schema,
    (value): IssueRewardCommand => {
      const p = fields(value);
      if (typeof p.benefit !== 'string' || !Number.isInteger(p.validDays) || Number(p.validDays) < 1)
        throw new Error('Invalid grant');
      return {
        grantId: identifier(String(p.grantId)),
        accountId: identifier(String(p.accountId)),
        benefit: p.benefit,
        validDays: Number(p.validDays),
      };
    },
  );
