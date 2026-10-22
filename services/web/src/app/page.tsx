'use client';
import { useCallback, useEffect, useState } from 'react';
import { useCafe } from '../shared/realtime/useCafe';
import { RealtimeProvider } from '../shared/realtime';
import {
  login as loginSession,
  logout as logoutSession,
  session as loadSession,
} from '../adaptors/generated/http/session/session';
import { httpStatus } from '../adaptors/http/operation-client';
import { useCommand } from '../shared/commands/useCommand';
import { MenuPanel } from '../features/menu/MenuPanel';
import { OrderingPanel } from '../features/ordering/OrderingPanel';
import { PreparationPanel } from '../features/preparation/PreparationPanel';
import { CollectionPanel } from '../features/collection/CollectionPanel';
import { RewardsPanel } from '../features/rewards/RewardsPanel';
import { NotificationsPanel } from '../features/notifications/NotificationsPanel';

export default function Cafe() {
  const [principalId, setPrincipalId] = useState('');
  const [checking, setChecking] = useState(true);
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [customer, setCustomer] = useState('');
  const sessionEnded = useCallback(() => setPrincipalId(''), []);
  useEffect(() => {
    void loadSession()
      .then((response) => {
        setPrincipalId(response.data.subject);
      })
      .catch((error: unknown) => {
        if (httpStatus(error) !== 401)
          setLoginError('The café API is unavailable. Try opening the café again.');
      })
      .finally(() => setChecking(false));
    let identity = localStorage.getItem('cafe:customer');
    if (!identity) {
      identity = crypto.randomUUID();
      localStorage.setItem('cafe:customer', identity);
    }
    setCustomer(identity);
  }, []);
  async function login(event: React.FormEvent) {
    event.preventDefault();
    setLoginError('');
    try {
      await loginSession({ password });
      setPassword('');
      const current = await loadSession();
      setPrincipalId(current.data.subject);
    } catch (error) {
      setLoginError(
        [401, 403].includes(httpStatus(error) ?? 0)
          ? 'The access code was not accepted.'
          : 'The café API is unavailable.',
      );
    }
  }
  if (checking)
    return (
      <main className="login">
        <p>Opening the café…</p>
      </main>
    );
  if (!principalId)
    return (
      <main className="login">
        <div className="brand">◒ CAFÉ LAB</div>
        <h1>
          A small café.
          <br />A complete journey.
        </h1>
        <p>Publish a menu, prepare an order and watch a reward arrive.</p>
        <form onSubmit={(event) => void login(event)}>
          <label>
            Local operator access code
            <input
              autoFocus
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          <button type="submit">Open the café →</button>
          <p role="alert">{loginError}</p>
        </form>
        <small>Development reference · Go / Python / TypeScript</small>
      </main>
    );
  return (
    <RealtimeProvider principalId={principalId} sessionEnded={sessionEnded}>
      <CafeWorkspace
        principalId={principalId}
        customer={customer}
        setCustomer={setCustomer}
        sessionEnded={sessionEnded}
        signOut={() => setPrincipalId('')}
      />
    </RealtimeProvider>
  );
}

function CafeWorkspace({
  principalId,
  customer,
  setCustomer,
  sessionEnded,
  signOut,
}: {
  principalId: string;
  customer: string;
  setCustomer: (customer: string) => void;
  sessionEnded: () => void;
  signOut: () => void;
}) {
  const { snapshot, connection, problem } = useCafe(principalId);
  const command = useCommand(sessionEnded);
  const disabled = command.busy || Boolean(command.pending) || connection !== 'connected';
  return (
    <main>
      <header>
        <div>
          <div className="brand">◒ CAFÉ LAB</div>
          <h1>From menu to reward.</h1>
          <p>One café. Six contexts. Follow each order as the work moves forward.</p>
        </div>
        <div className="connection">
          <span className={'indicator ' + connection} />
          <span data-testid="connection">
            {connection === 'connected' ? 'Live updates connected' : 'Reconnecting…'}
          </span>
          <button className="quiet" onClick={() => window.dispatchEvent(new Event('cafe-reconnect'))}>
            Reconnect
          </button>
          <button
            className="quiet"
            onClick={() => {
              void logoutSession()
                .then(() => signOut())
                .catch(() => sessionEnded());
            }}
          >
            Sign out
          </button>
        </div>
      </header>
      <div className="service-strip">
        <span>
          Storefront <b>Go</b>
        </span>
        <span>
          Operations <b>Python</b>
        </span>
        <span>
          Engagement <b>TypeScript</b>
        </span>
      </div>
      <section className="customer-bar">
        <label>
          Customer identity
          <input
            aria-label="Customer identity"
            value={customer}
            onChange={(e) => {
              setCustomer(e.target.value);
              localStorage.setItem('cafe:customer', e.target.value);
            }}
          />
        </label>
        <button
          className="quiet"
          onClick={() => {
            const id = crypto.randomUUID();
            setCustomer(id);
            localStorage.setItem('cafe:customer', id);
          }}
        >
          New customer
        </button>
        <p>Every three collected orders earns one free drink.</p>
      </section>
      {(command.message || problem) && (
        <aside className="feedback" role="status">
          {problem || command.message}
          {command.pending && !command.busy && (
            <button onClick={() => void command.retry()}>Retry the same command</button>
          )}
        </aside>
      )}
      <div className="workspace">
        <MenuPanel snapshot={snapshot} send={command.send} disabled={disabled} />
        <OrderingPanel snapshot={snapshot} customer={customer} send={command.send} disabled={disabled} />
        <section className="panel preparation">
          <div className="section-title">
            <span>03 / OPERATIONS</span>
            <h2>Behind the counter</h2>
          </div>
          <PreparationPanel snapshot={snapshot} send={command.send} disabled={disabled} />
          <CollectionPanel snapshot={snapshot} send={command.send} disabled={disabled} />
        </section>
      </div>
      <section className="engagement">
        <RewardsPanel snapshot={snapshot} customer={customer} send={command.send} disabled={disabled} />
        <NotificationsPanel snapshot={snapshot} customer={customer} />
      </section>
      <footer>
        Development café · Authoritative state stays with its owning context · Live delivery via Centrifugo
      </footer>
    </main>
  );
}
