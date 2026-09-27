'use client';
import {useCallback, useEffect, useState} from 'react';
import {useCafe} from '../features/realtime/useCafe';
import {items} from '../features/realtime/model';
import {useCommand} from '../features/commands/useCommand';

const short = (id: string) => id.slice(0, 8);
const money = (minor: number, currency = 'EUR') => new Intl.NumberFormat('en-GB', {style: 'currency', currency}).format(minor / 100);

export default function Cafe() {
  const [authenticated, setAuthenticated] = useState(false);
  const [checking, setChecking] = useState(true);
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState('');
  const [customer, setCustomer] = useState('');
  const [drinkName, setDrinkName] = useState('Cappuccino');
  const [editionId, setEditionId] = useState('');
  const [drinkId, setDrinkId] = useState('');
  const [offerCode, setOfferCode] = useState('C1');
  const [price, setPrice] = useState('3.00');
  const [quantity, setQuantity] = useState(1);
  const [codes, setCodes] = useState<Record<string, string>>({});
  const sessionEnded = useCallback(() => setAuthenticated(false), []);
  const {snapshot, connection, problem} = useCafe(authenticated, sessionEnded);
  const command = useCommand(sessionEnded);
  const drinks = items(snapshot, 'drink'), editions = items(snapshot, 'edition'), orders = items(snapshot, 'order');
  const tickets = items(snapshot, 'ticket'), pickups = items(snapshot, 'pickup');
  const accounts = items(snapshot, 'account'), rewards = items(snapshot, 'reward'), notices = items(snapshot, 'notification');
  const selectedEdition = editions.find(item => item.state.id === editionId) ?? editions[0];
  const published = editions.filter(item => item.state.status === 'published');
  const disabled = command.busy || Boolean(command.pending) || connection !== 'connected';
  useEffect(() => {
    void fetch('/auth/session', {cache: 'no-store'}).then(response => setAuthenticated(response.ok))
      .catch(() => setLoginError('The café API is unavailable. Try opening the café again.'))
      .finally(() => setChecking(false));
    let identity = localStorage.getItem('cafe:customer');
    if (!identity) { identity = crypto.randomUUID(); localStorage.setItem('cafe:customer', identity); }
    setCustomer(identity);
  }, []);
  async function login(event: React.FormEvent) {
    event.preventDefault(); setLoginError('');
    try {
      const response = await fetch('/auth/login', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({password})});
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
          const response = await fetch('/auth/logout', {method: 'POST'});
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
      <section className="panel menu"><div className="section-title"><span>01 / STOREFRONT</span><h2>The menu</h2></div>
        <div className="form-row"><label>Drink name<input value={drinkName} onChange={e => setDrinkName(e.target.value)}/></label>
          <button disabled={disabled || !drinkName.trim()} onClick={() => void command.send('menu/drinks/'+crypto.randomUUID(), {name: drinkName}, 0)}>Create drink</button></div>
        {drinks.map(item => <div className="list-row" key={item.state.id}><div><strong>{item.state.name}</strong><small>Revision {item.state.revision || 'draft'}</small></div>
          {item.state.published ? <span className="badge">Published</span> :
            <button disabled={disabled} onClick={() => void command.send(`menu/drinks/${item.state.id}/publish`, {}, item.version)}>Publish drink</button>}</div>)}
        <div className="divider"/>
        <div className="form-row"><label>Menu edition<select value={selectedEdition?.state.id ?? ''} onChange={e => setEditionId(e.target.value)}>
          <option value="" disabled>Select an edition</option>{editions.map(item => <option key={item.state.id} value={item.state.id}>{short(item.state.id)} · {item.state.status}</option>)}
        </select></label><button disabled={disabled} onClick={async () => {
          const id = crypto.randomUUID(); if (await command.send('menu/editions/'+id, {currency: 'EUR'}, 0)) setEditionId(id);
        }}>New edition</button></div>
        {selectedEdition && <>
          {selectedEdition.state.offers.map(offer => <div className="list-row" key={offer.code}><span>{offer.code} · {offer.name}</span><strong>{money(offer.minor, offer.currency)}</strong></div>)}
          {selectedEdition.state.status === 'draft' ? <div className="offer-form"><label>Published drink<select value={drinkId} onChange={e => setDrinkId(e.target.value)}>
            <option value="">Choose a drink</option>{drinks.filter(d => d.state.published).map(d => <option key={d.state.id} value={d.state.id}>{d.state.name}</option>)}</select></label>
            <div className="form-row"><label>Code<input value={offerCode} onChange={e => setOfferCode(e.target.value.toUpperCase())}/></label>
              <label>Price (€)<input inputMode="decimal" value={price} onChange={e => setPrice(e.target.value)}/></label></div>
            <div className="form-row"><button disabled={disabled || !drinkId || !/^\d+(\.\d{1,2})?$/.test(price)} onClick={() => {
              const [whole, fraction = ''] = price.split('.'); const drink = drinks.find(d => d.state.id === drinkId)!;
              void command.send(`menu/editions/${selectedEdition.state.id}/offers`, {code: offerCode, drinkId,
                drinkRevision: drink.state.revision, minor: Number(whole)*100+Number(fraction.padEnd(2, '0'))}, selectedEdition.version);
            }}>Add offer</button><button disabled={disabled || selectedEdition.state.offers.length === 0}
              onClick={() => void command.send(`menu/editions/${selectedEdition.state.id}/publish`, {}, selectedEdition.version)}>Publish menu</button></div>
          </div> : <p className="note">This edition is published. Its names and prices are frozen.</p>}
        </>}
      </section>
      <section className="panel orders"><div className="section-title"><span>02 / ORDERING</span><h2>At the counter</h2></div>
        <label>Drinks per line<input type="number" min="1" max="5" value={quantity} onChange={e => setQuantity(Number(e.target.value))}/></label>
        <button className="wide" disabled={disabled || published.length === 0 || !customer} onClick={() =>
          void command.send('ordering/orders/'+crypto.randomUUID(), {customerId: customer, editionId: published.at(-1)!.state.id}, 0)}>Start an order</button>
        {!published.length && <p className="empty">Publish a menu to open the counter.</p>}
        {orders.filter(order => order.state.customerId === customer).map(order => <article className="card" data-testid="order" key={order.state.id}>
          <div className="card-heading"><strong>Order {short(order.state.id)}</strong><span className="badge">{order.state.status}</span></div>
          {order.state.lines.map(line => <p key={line.id}>{line.quantity} × {line.selection.name} <span>{money(line.selection.minor*line.quantity, order.state.currency)}</span></p>)}
          {order.state.status === 'draft' ? <><div className="offer-buttons">
            {editions.find(e => e.state.id === order.state.editionId)?.state.offers.map(offer => <button className="quiet" disabled={disabled} key={offer.code}
              onClick={() => void command.send(`ordering/orders/${order.state.id}/lines`, {lineId: crypto.randomUUID(),
                editionId: order.state.editionId, offerCode: offer.code, quantity}, order.version)}>+ {offer.name}</button>)}</div>
            <button className="wide" disabled={disabled || order.state.lines.length === 0}
              onClick={() => void command.send(`ordering/orders/${order.state.id}/place`, {}, order.version)}>Place order</button></> :
            <small>{pickups.some(p => p.state.orderId === order.state.id && p.state.status === 'collected') ? 'Collected' :
              tickets.some(t => t.state.orderId === order.state.id) ? 'With the preparation team' : 'Waiting for preparation'}</small>}
        </article>)}
      </section>
      <section className="panel preparation"><div className="section-title"><span>03 / OPERATIONS</span><h2>Behind the counter</h2></div>
        <h3>Preparation</h3>
        {!tickets.length && <p className="empty">Placed orders arrive here automatically.</p>}
        {tickets.filter(t => t.state.status !== 'ready').map(ticket => <article className="card" data-testid="ticket" key={ticket.state.id}>
          <div className="card-heading"><strong>Order {short(ticket.state.orderId)}</strong><span className="badge">{ticket.state.status}</span></div>
          <p>{ticket.state.instructions}</p><button disabled={disabled} onClick={() => void command.send(
            `preparation/tickets/${ticket.state.id}/${ticket.state.status === 'queued' ? 'start' : 'complete'}`, {}, ticket.version)}>
            {ticket.state.status === 'queued' ? 'Start preparation' : 'Mark drinks ready'}</button>
        </article>)}
        <h3>Collection</h3>
        {pickups.filter(p => p.state.status === 'ready').map(pickup => <article className="card" data-testid="pickup" key={pickup.state.id}>
          <div className="card-heading"><strong>Order {short(pickup.state.orderId)}</strong><span className="badge">Ready</span></div>
          <p>Collection code <b data-testid="collection-code">{pickup.state.code}</b></p>
          <div className="form-row"><input aria-label={'Code for '+short(pickup.state.orderId)} placeholder="Enter collection code"
            value={codes[pickup.state.id] ?? ''} onChange={e => setCodes({...codes, [pickup.state.id]: e.target.value.toUpperCase()})}/>
            <button disabled={disabled} onClick={() => void command.send(`collection/pickups/${pickup.state.id}/collect`,
              {code: codes[pickup.state.id] ?? ''}, pickup.version)}>Collect order</button></div>
        </article>)}
      </section>
    </div>
    <section className="engagement"><div><div className="section-title"><span>04 / ENGAGEMENT</span><h2>A little thank you.</h2></div>
      {accounts.filter(a => a.state.id === customer).map(account => <div key={account.state.id} data-testid="loyalty-account">
        <div className="stamps">{[0, 1, 2].map(i => <span className={i < account.state.stampBalance ? 'filled' : ''} key={i}>◒</span>)}</div>
        <p>{account.state.collections} orders collected · {account.state.grantsEarned} rewards earned</p>
        {account.state.grantsEarned > rewards.filter(r => r.state.customerId === customer).length &&
          <p>A grant is earned. Reward issuance is pending.</p>}
      </div>)}
      {rewards.filter(r => r.state.customerId === customer).map(reward => <article className="reward" data-testid="reward" key={reward.state.id}>
        <strong>{reward.state.benefit}</strong><span className="badge">{reward.state.status}</span>
        <small>Valid until {new Date(reward.state.expiresAt).toLocaleDateString('en-GB')}</small>
        {reward.state.status === 'issued' && <button className="quiet" disabled={disabled || orders.length === 0}
          onClick={() => void command.send(`loyalty/rewards/${reward.state.id}/redeem`, {orderId: orders.filter(o => o.state.customerId === customer).at(-1)?.state.id ?? ''}, reward.version)}>Record redemption</button>}
      </article>)}</div>
      <div><h3>Customer messages</h3>{notices.filter(n => n.state.recipient === customer).map(notice => <article className="message" data-testid="notification" key={notice.state.id}>
        <div className="card-heading"><strong>{notice.state.subject}</strong><span className="badge">{notice.state.status}</span></div><p>{notice.state.body}</p></article>)}
        <p className="note">Messages go to the local mailbox simulator. No email is sent.</p></div>
    </section>
    <footer>Development café · Authoritative state stays with its owning context · Live delivery via Centrifugo</footer>
  </main>;
}
