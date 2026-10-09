import {createServer, type ServerResponse} from 'node:http';
import {timingSafeEqual} from 'node:crypto';

function json(response: ServerResponse, status: number, body: unknown) {
  response.writeHead(status, {'content-type': 'application/json'});
  response.end(JSON.stringify(body));
}

/** Exposes health and authenticated diagnostics for the owner process. */
export function server(key: string, diagnostics: () => Promise<unknown>) {
  if (key.length < 32) throw new Error('A development service key is required');
  return createServer((request, response) => {
    void (async () => {
      const path = new URL(request.url ?? '/', 'http://internal').pathname;
      if (path === '/healthz' && request.method === 'GET') {
        json(response, 200, {status: 'ok', language: 'typescript'}); return;
      }
      const supplied = Buffer.from(request.headers.authorization ?? ''), expected = Buffer.from('Bearer '+key);
      if (supplied.length !== expected.length || !timingSafeEqual(supplied, expected)) {
        json(response, 401, {code: 'unauthorised'}); return;
      }
      if (path === '/diagnostics' && request.method === 'GET') {
        response.setHeader('cache-control', 'no-store'); json(response, 200, await diagnostics()); return;
      }
      json(response, 404, {code: 'not_found'});
    })().catch(() => json(response, 503, {code: 'temporarily_unavailable'}));
  });
}
