import {routes, type Operation, type Input, type Success, type CommandOperation, type CommandBody, type CommandParameters, type Intersection} from '../generated/http';

/** Signals that the HTTP session requires authentication before a retained command can be retried. */
export class SessionEnded extends Error {
  constructor() { super('Session ended'); }
}
// A caller with a union of operations must satisfy every possible input contract.
export async function request<K extends Operation>(operation: K, input: Input<NoInfer<K>> & Intersection<Input<NoInfer<K>>>,
  options: {signal?: AbortSignal} = {}): Promise<Response> {
  return send(routes[operation], input, options.signal);
}
type HTTPInput = {path?: object; query?: object; headers?: object; body?: unknown};
function send(route: {path: string; method: string}, input: HTTPInput, signal?: AbortSignal): Promise<Response> {
  let url: string = route.path;
  if (input.path) for (const [name, value] of Object.entries(input.path)) {
    url = url.replace('{'+name+'}', encodeURIComponent(String(value)));
  }
  if (url.includes('{')) throw new Error('HTTP path parameter is missing');
  if (input.query) {
    const query = new URLSearchParams();
    for (const [name, value] of Object.entries(input.query)) {
      if (value !== undefined) query.set(name, String(value));
    }
    if (query.size) url += '?'+query;
  }
  const headers = new Headers();
  for (const [name, value] of Object.entries(input.headers ?? {})) {
    if (value !== undefined) headers.set(name, String(value));
  }
  if (input.body !== undefined) headers.set('Content-Type', 'application/json');
  return fetch(url, {method: route.method, headers, cache: 'no-store', signal,
    body: input.body === undefined ? undefined : JSON.stringify(input.body)});
}
export async function success<K extends Operation>(_operation: K, response: Response): Promise<Success<K>> {
  if (response.status === 401) throw new SessionEnded();
  if (response.status !== 200) throw new Error('The current café state could not be loaded. Reconnect to retry.');
  return response.json();
}

// The shared command hook must satisfy the parameters of every operation it sends.
// An added required parameter therefore breaks the hook's actual call site.
export function commandRequest<K extends CommandOperation>(operation: K,
  body: CommandBody<NoInfer<K>>, parameters: CommandParameters): Promise<Response> {
  return send(routes[operation], {...parameters, body});
}
