'use client';
import {useCallback, useEffect, useState} from 'react';
import {useCafe} from '../features/realtime/useCafe';
import {request} from '../adaptors/http/client';
import {useCommand} from '../features/commands/useCommand';
import {MenuPanel} from '../features/cafe/MenuPanel';
import {OrderingPanel} from '../features/cafe/OrderingPanel';
import {OperationsPanel} from '../features/cafe/OperationsPanel';
import {EngagementPanel} from '../features/cafe/EngagementPanel';

export default function Cafe() {
  const [authenticated, setAuthenticated] = useState(false);
  const [checking, setChecking] = useState(true);
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [customer, setCustomer] = useState('');
  const sessionEnded = useCallback(() => setAuthenticated(false), []);
  const {snapshot, connection, problem} = useCafe(authenticated, sessionEnded);
  const command = useCommand(sessionEnded);
  const disabled = command.busy || Boolean(command.pending) || connection !== 'connected';
  useEffect(() => {
    void request('session', {}).then(response => setAuthenticated(response.ok))
      .catch(() => setLoginError('The café API is unavailable. Try opening the café again.'))
      .finally(() => setChecking(false));
    let identity = localStorage.getItem('cafe:customer');
    if (!identity) { identity = crypto.randomUUID(); localStorage.setItem('cafe:customer', identity); }
    setCustomer(identity);
  }, []);
  async function login(event: React.FormEvent) {
    event.preventDefault(); setLoginError('');
    try {
      const response = await request('login', {body: {password}});
      if (!response.ok) { setLoginError('The access code was not accepted.'); return; }
      setPassword(''); setAuthenticated(true);
    } catch { setLoginError('The café API is unavailable.'); }
  }
  if (checking) return <main className="login"><p>Opening the café…</p></main>;
  if (!authenticated) return <main className="login"><div className="brand">◒ CAFÉ LAB</div>
    <h1>A small café.<br/>A complete journey.</h1><p>Publish a menu, prepare an order and watch a reward arrive.</p>
    <form onSubmit={login}><label>Local operator access code<input autoFocus type="password" value={password}
      onChange={e => setPassword(e.target.value)} autoComplete="current-password" required/></label>
      <button type="submit">Open the café →</button><p role="alert">{loginError}</p></form>
    <small>Development reference · Go / Python / TypeScript</small></main>;
  return <main>
    <header><div><div className="brand">◒ CAFÉ LAB</div><h1>From menu to reward.</h1>
      <p>One café. Six contexts. Follow each order as the work moves forward.</p></div>
      <div className="connection"><span className={'indicator '+connection}/><span data-testid="connection">{connection === 'connected' ? 'Live updates connected' : 'Reconnecting…'}</span>
        <button className="quiet" onClick={() => window.dispatchEvent(new Event('cafe-reconnect'))}>Reconnect</button>
        <button className="quiet" onClick={async () => {
          const response = await request('logout', {});
          if (response.ok) setAuthenticated(false);
        }}>Sign out</button></div>
    </header>
    <div className="service-strip"><span>Storefront <b>Go</b></span><span>Operations <b>Python</b></span><span>Engagement <b>TypeScript</b></span></div>
    <section className="customer-bar"><label>Customer identity<input aria-label="Customer identity" value={customer} onChange={e => {
      setCustomer(e.target.value); localStorage.setItem('cafe:customer', e.target.value);
    }}/></label><button className="quiet" onClick={() => {
      const id = crypto.randomUUID(); setCustomer(id); localStorage.setItem('cafe:customer', id);
    }}>New customer</button><p>Every three collected orders earns one free drink.</p></section>
    {(command.message || problem) && <aside className="feedback" role="status">
      {problem || command.message}{command.pending && !command.busy && <button onClick={() => void command.retry()}>Retry the same command</button>}
    </aside>}
    <div className="workspace">
      <MenuPanel snapshot={snapshot} send={command.send} disabled={disabled}/>
      <OrderingPanel snapshot={snapshot} customer={customer} send={command.send} disabled={disabled}/>
      <OperationsPanel snapshot={snapshot} send={command.send} disabled={disabled}/>
    </div>
    <EngagementPanel snapshot={snapshot} customer={customer} send={command.send} disabled={disabled}/>
    <footer>Development café · Authoritative state stays with its owning context · Live delivery via Centrifugo</footer>
  </main>;
}
