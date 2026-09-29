import {Rejection} from '../foundation/domain.js';

export function redeemInput(value: unknown): {orderId: string} {
  if (!value || typeof value !== 'object' || Array.isArray(value) ||
      Object.keys(value).length !== 1 || !('orderId' in value) || typeof value.orderId !== 'string') {
    throw new Rejection('invalid_request', 'An orderId string is required');
  }
  return {orderId: value.orderId};
}
