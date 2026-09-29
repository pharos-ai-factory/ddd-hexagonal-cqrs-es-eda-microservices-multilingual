import {createServer, type IncomingMessage, type ServerResponse} from 'node:http';
import {timingSafeEqual} from 'node:crypto';
import type {Metadata, Outcome, QueryPort} from '../foundation/application.js';
import type {PagedQueryPort} from '../foundation/pagination.js';
import {pageRequest, pageResponse} from './pagination.js';
import {identifier, Rejection} from '../foundation/domain.js';
type Route = {
  resource: string; queries: QueryPort<unknown> & Partial<Pick<PagedQueryPort<unknown>, 'page'>>;
  commands?: Record<string, {name: string; invoke(m: Omit<Metadata, 'input'>, value: unknown): Promise<Outcome>}>;
};
function json(response: ServerResponse, status: number, body: unknown) {
  response.writeHead(status, {'content-type': 'application/json'});
  response.end(JSON.stringify(body));
}
async function body(request: IncomingMessage) {
  const buffers: Buffer[] = []; let length = 0;
  for await (const chunk of request) {
    length += chunk.length;
    if (length > 65536) throw new Rejection('invalid_request', 'Command body is too large');
    buffers.push(chunk);
  }
  return JSON.parse(Buffer.concat(buffers).toString()) as unknown;
}
export function server(key: string, routes: Route[], diagnostics?: () => Promise<unknown>) {
  if (key.length < 32) throw new Error('A development service key is required');
  return createServer((request, response) => {
    void (async () => {
      if (request.url === '/healthz') { json(response, 200, {status: 'ok', language: 'typescript'}); return; }
      const supplied = Buffer.from(request.headers.authorization ?? ''), expected = Buffer.from('Bearer '+key);
      if (supplied.length !== expected.length || !timingSafeEqual(supplied, expected)) {
        json(response, 401, {code: 'unauthorised'}); return;
      }
      const url = new URL(request.url ?? '/', 'http://internal'), path = url.pathname;
      if (path === '/diagnostics' && request.method === 'GET' && diagnostics) {
        response.setHeader('cache-control', 'no-store'); json(response, 200, await diagnostics()); return;
      }
      for (const route of routes) {
        if (path === route.resource && request.method === 'GET') {
          const page = pageRequest(url.searchParams, path);
          if (page && !route.queries.page) throw new Rejection('invalid_pagination', 'This query does not support pagination');
          json(response, 200, page ? pageResponse(await route.queries.page!(page), path) : await route.queries.list()); return;
        }
        if (!path.startsWith(route.resource+'/')) continue;
        const [identity, action, extra] = path.slice(route.resource.length+1).split('/');
        identifier(identity ?? '');
        if (extra) break;
        if (!action && request.method === 'GET') {
          const loaded = await route.queries.get(identity!);
          json(response, loaded ? 200 : 404, loaded ?? {code: 'not_found'}); return;
        }
        const handler = action ? route.commands?.[action] : undefined;
        if (!handler || request.method !== 'POST') break;
        const input = await body(request);
        const versionHeader = String(request.headers['if-match'] ?? '').replace(/^"|"$/g, '');
        const version = Number(versionHeader);
        if (!/^\d+$/.test(versionHeader) || !Number.isSafeInteger(version)) {
          json(response, 428, {code: 'expected_version_required'}); return;
        }
        const commandId = identifier(String(request.headers['idempotency-key'] ?? ''));
        const m: Omit<Metadata, 'input'> = {id: commandId, target: identity!, name: handler.name, expected: version,
          correlation: identifier(String(request.headers['x-correlation-id'] ?? commandId))};
        const outcome = await handler.invoke(m, input);
        const code = outcome.rejection?.code;
        json(response, code === 'version_conflict' || code === 'idempotency_conflict' ? 409 :
          code === 'not_found' ? 404 : code ? 422 : 200, outcome);
        return;
      }
      json(response, 404, {code: 'not_found'});
    })().catch(error => json(response, error instanceof Rejection || error instanceof SyntaxError ? 400 : 503,
      error instanceof Rejection ? error.outcome() : {code: 'temporarily_unavailable'}));
  });
}
