import {CorruptState, identifier} from '../../../foundation/domain.js';

export type Grant = Readonly<{id: string; accountId: string; benefit: string; validDays: number}>;
export type AccountState = Readonly<{
  id: string; stampBalance: number; collections: number; grantsEarned: number; lastGrant?: Grant;
}>;
export type AccountFact = Readonly<{type: 'RewardEarned'; grant: Grant} | {type: 'StampCredited'; orderId: string}>;

/** Owns collection credits, stamp balance and immutable earned reward grants. */
export class LoyaltyAccount {
  #state: AccountState;
  #events: AccountFact[] = [];
  constructor(state: AccountState) {
    try { identifier(state.id); }
    catch (cause) { throw new CorruptState('Corrupt account identity', {cause}); }
    if (!Number.isSafeInteger(state.collections) || state.collections < 0 ||
        !Number.isSafeInteger(state.grantsEarned) || state.grantsEarned < 0 ||
        ![0, 1, 2].includes(state.stampBalance) ||
        state.collections !== state.grantsEarned * 3 + state.stampBalance) {
      throw new CorruptState('Corrupt stamp accounting');
    }
    if ((state.grantsEarned > 0 && !state.lastGrant) || (state.grantsEarned === 0 && state.lastGrant)) {
      throw new CorruptState('Corrupt earned grant');
    }
    if (state.lastGrant) {
      try {
        identifier(state.lastGrant.id);
        identifier(state.lastGrant.accountId);
      } catch (cause) { throw new CorruptState('Corrupt earned grant', {cause}); }
      if (state.lastGrant.accountId !== state.id || typeof state.lastGrant.benefit !== 'string' ||
          !state.lastGrant.benefit || !Number.isSafeInteger(state.lastGrant.validDays) ||
          state.lastGrant.validDays < 1 || state.lastGrant.validDays > 30) {
        throw new CorruptState('Corrupt earned grant');
      }
    }
    this.#state = structuredClone(state);
  }
  static open(id: string) {
    identifier(id);
    return new LoyaltyAccount({id, collections: 0, grantsEarned: 0, stampBalance: 0});
  }

  // The application receipt owns the collection business key; history is not loaded into this root.
  credit(orderId: string, grantId: string): void {
    identifier(orderId); identifier(grantId);
    let state = {...this.#state, collections: this.#state.collections + 1, stampBalance: this.#state.stampBalance + 1};
    this.#events.push({type: 'StampCredited', orderId});
    if (state.stampBalance === 3) {
      const grant = Object.freeze({id: grantId, accountId: state.id, benefit: 'one free drink', validDays: 7});
      state = {...state, stampBalance: 0, grantsEarned: state.grantsEarned + 1, lastGrant: grant};
      this.#events.push({type: 'RewardEarned', grant});
    }
    this.#state = state;
  }
  snapshot(): AccountState { return structuredClone(this.#state); }
  events(): readonly AccountFact[] { return structuredClone(this.#events); }
}
