import {CorruptState, DomainError, identifier, Rejection} from '../../../foundation/domain.js';
import type {Grant} from './loyalty-account.js';

/** Identifies refusal to redeem a reward outside its valid lifecycle. */
export class RewardUnavailableDomainError extends DomainError {
  constructor() { super('reward_unavailable', 'The reward is no longer available'); }
}

export type RewardState = Readonly<{
  id: string; grantId: string; customerId: string; benefit: string;
  status: 'issued' | 'redeemed' | 'expired'; expiresAt: string; redeemedFor?: string;
}>;
/** Owns issuance, expiry and redemption of one earned reward. */
export class Reward {
  #events: {type: 'RewardIssued'; reward: RewardState}[] = [];
  #state: RewardState;
  constructor(state: RewardState) {
    try {
      [state.id, state.grantId, state.customerId].forEach(identifier);
      if (!['issued', 'redeemed', 'expired'].includes(state.status) || !Number.isFinite(Date.parse(state.expiresAt)) ||
          typeof state.benefit !== 'string' || !state.benefit) throw new Error('Invalid reward terms');
      if (state.status === 'redeemed') identifier(state.redeemedFor!);
      else if (state.redeemedFor !== undefined) throw new Error('Unexpected redemption');
    } catch (cause) {
      throw new CorruptState('Corrupt reward state', {cause});
    }
    this.#state = {...state};
  }
  static issue(id: string, grant: Grant, now: Date): Reward {
    [id, grant.id, grant.accountId].forEach(identifier);
    if (!grant.benefit || !Number.isInteger(grant.validDays) || grant.validDays < 1 || grant.validDays > 30) {
      throw new Rejection('invalid_grant', 'The earned grant has invalid terms');
    }
    const reward = new Reward({id, grantId: grant.id, customerId: grant.accountId, benefit: grant.benefit,
      status: 'issued', expiresAt: new Date(now.getTime() + grant.validDays * 86400000).toISOString()});
    reward.#events.push({type: 'RewardIssued', reward: reward.snapshot()});
    return reward;
  }
  redeem(orderId: string, now: Date): void {
    identifier(orderId);
    if (this.#state.status !== 'issued') throw new RewardUnavailableDomainError();
    if (now.getTime() >= Date.parse(this.#state.expiresAt)) throw new Rejection('reward_expired', 'The reward has expired');
    this.#state = {...this.#state, status: 'redeemed', redeemedFor: orderId};
  }
  expire(now: Date): void {
    if (this.#state.status !== 'issued') throw new RewardUnavailableDomainError();
    if (now.getTime() < Date.parse(this.#state.expiresAt)) throw new Rejection('reward_not_expired', 'Validity has not ended');
    this.#state = {...this.#state, status: 'expired'};
  }
  events(): readonly {type: 'RewardIssued'; reward: RewardState}[] { return structuredClone(this.#events); }
  snapshot(): RewardState { return {...this.#state}; }
}
