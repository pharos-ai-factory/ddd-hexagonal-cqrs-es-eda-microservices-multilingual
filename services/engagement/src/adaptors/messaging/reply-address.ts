/** Validate API connection reply addresses within the receiving context. */
export function replyRoute(owner: string, address: unknown): string {
  const legacy = 'reply.' + owner;
  if (address === undefined || address === null || address === '' || address === legacy) return legacy;
  if (
    typeof address !== 'string' ||
    !address.startsWith(legacy + '.') ||
    !/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(
      address.slice(legacy.length + 1),
    )
  )
    throw new Error('Invalid owner reply address');
  return address;
}
