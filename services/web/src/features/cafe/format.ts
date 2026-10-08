export type {SendCommand} from '../../adaptors/generated/http';

export const short = (id: string) => id.slice(0, 8);
export const money = (minor: number, currency = 'EUR') =>
  new Intl.NumberFormat('en-GB', {style: 'currency', currency}).format(minor / 100);

export function minorUnits(price: string): number | undefined {
  if (!/^\d+(\.\d{1,2})?$/.test(price)) return undefined;
  const [whole, fraction = ''] = price.split('.');
  const minor = Number(whole) * 100 + Number(fraction.padEnd(2, '0'));
  return Number.isSafeInteger(minor) ? minor : undefined;
}
