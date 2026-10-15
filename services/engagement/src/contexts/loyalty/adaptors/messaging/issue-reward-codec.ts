import { LoyaltyPrivateCommand, LoyaltySubscription } from '../../../../adaptors/generated/message-names.js';
import { parseIssueRewardCommand } from './generated/request-payloads.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { IssueRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';

export const issueCodec = () =>
  new InternalCommandCodec(
    'loyalty',
    LoyaltySubscription.ISSUE_REWARD,
    LoyaltyPrivateCommand.ISSUE_REWARD,
    schema,
    (value): IssueRewardCommand => {
      const command = parseIssueRewardCommand(value);
      identifier(command.grantId);
      identifier(command.accountId);
      if (command.validDays < 1) throw new Error('Invalid grant');
      return command;
    },
  );
