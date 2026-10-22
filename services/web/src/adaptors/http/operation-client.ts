import { operations, type OperationName, type OperationTypes } from '../generated/http/operations';

type BusinessTag = 'menu' | 'ordering' | 'preparation' | 'collection' | 'loyalty' | 'communication';
type BusinessOperation<Method extends string> = {
  [Name in OperationName]: (typeof operations)[Name]['method'] extends Method
    ? (typeof operations)[Name]['tag'] extends BusinessTag
      ? Name
      : never
    : never;
}[OperationName];

export type CommandOperation = BusinessOperation<'POST'>;
export type ListOperation = Extract<BusinessOperation<'GET'>, `list${string}`>;
export type CommandBody<Name extends CommandOperation> = OperationTypes[Name]['body'];
export type SendCommand = <Name extends CommandOperation>(
  operation: Name,
  id: string,
  body: CommandBody<NoInfer<Name>>,
  version: number,
) => Promise<boolean>;

type Intersection<Union> = (Union extends unknown ? (value: Union) => void : never) extends (
  value: infer Result,
) => void
  ? Result
  : never;
type Parameter<Name extends string, Value> = [Value] extends [never]
  ? { [Key in Name]?: never }
  : object extends Value
    ? { [Key in Name]?: Value }
    : { [Key in Name]: Value };
type CommandParameterValue<Key extends 'pathParams' | 'query' | 'headers'> = Intersection<
  { [Name in CommandOperation]: OperationTypes[Name][Key] }[CommandOperation]
>;
type CommandParameters = Parameter<'path', CommandParameterValue<'pathParams'>> &
  Parameter<'query', CommandParameterValue<'query'>> &
  Parameter<'headers', CommandParameterValue<'headers'>>;

/** Signals that the HTTP session requires authentication before a retained command can be retried. */
export class SessionEnded extends Error {
  constructor() {
    super('Session ended');
  }
}

/** Return the status attached to an error thrown by the generated Fetch client. */
export function httpStatus(error: unknown): number | undefined {
  if (!(error instanceof Error) || !('status' in error)) return undefined;
  const status = (error as Error & { status?: unknown }).status;
  return typeof status === 'number' ? status : undefined;
}

/** Send a durable business command while retaining its raw HTTP response for recovery decisions. */
export function commandRequest<Name extends CommandOperation>(
  operation: Name,
  body: CommandBody<NoInfer<Name>>,
  parameters: CommandParameters,
): Promise<Response> {
  const metadata = operations[operation];
  let url: string = metadata.path;
  for (const [name, value] of Object.entries(parameters.path)) {
    url = url.replace(`{${name}}`, encodeURIComponent(String(value)));
  }
  if (url.includes('{')) throw new Error('HTTP path parameter is missing');
  if (parameters.query) {
    const query = new URLSearchParams();
    for (const [name, value] of Object.entries(parameters.query)) {
      if (value !== undefined) query.set(name, String(value));
    }
    if (query.size) url += `?${query.toString()}`;
  }
  const headers = new Headers();
  for (const [name, value] of Object.entries(parameters.headers)) {
    if (value !== undefined) headers.set(name, String(value));
  }
  headers.set('Content-Type', 'application/json');
  return fetch(url, {
    method: metadata.method,
    headers,
    cache: 'no-store',
    body: JSON.stringify(body),
  });
}

/** Decode the successful outcome returned by a durable business command. */
export async function commandSuccess<Name extends CommandOperation>(
  _operation: Name,
  response: Response,
): Promise<OperationTypes[Name]['response']> {
  return (await response.json()) as OperationTypes[Name]['response'];
}

export { operations as routes };
