'use client';
import {useState} from 'react';
import {items, type Snapshot} from '../realtime/model';
import {short, type SendCommand} from './format';

export function OperationsPanel({snapshot, send, disabled}: {snapshot: Snapshot; send: SendCommand; disabled: boolean}) {
  const [codes, setCodes] = useState<Record<string, string>>({});
  const tickets = items(snapshot, 'ticket');
  const pickups = items(snapshot, 'pickup');

  return <section className="panel preparation"><div className="section-title"><span>03 / OPERATIONS</span><h2>Behind the counter</h2></div>
    <h3>Preparation</h3>
    {!tickets.length && <p className="empty">Placed orders arrive here automatically.</p>}
    {tickets.filter(t => t.state.status !== 'ready').map(ticket => <article className="card" data-testid="ticket" key={ticket.state.id}>
      <div className="card-heading"><strong>Order {short(ticket.state.orderId)}</strong><span className="badge">{ticket.state.status}</span></div>
      <p>{ticket.state.instructions}</p><button disabled={disabled} onClick={() => void send(
        ticket.state.status === 'queued' ? 'startPreparation' : 'completePreparation', ticket.state.id, {}, ticket.version)}>
        {ticket.state.status === 'queued' ? 'Start preparation' : 'Mark drinks ready'}</button>
    </article>)}
    <h3>Collection</h3>
    {pickups.filter(p => p.state.status === 'ready').map(pickup => <article className="card" data-testid="pickup" key={pickup.state.id}>
      <div className="card-heading"><strong>Order {short(pickup.state.orderId)}</strong><span className="badge">Ready</span></div>
      <p>Collection code <b data-testid="collection-code">{pickup.state.code}</b></p>
      <div className="form-row"><input aria-label={'Code for '+short(pickup.state.orderId)} placeholder="Enter collection code"
        value={codes[pickup.state.id] ?? ''} onChange={e => setCodes({...codes, [pickup.state.id]: e.target.value.toUpperCase()})}/>
        <button disabled={disabled} onClick={() => void send('collectOrder', pickup.state.id,
          {code: codes[pickup.state.id] ?? ''}, pickup.version)}>Collect order</button></div>
    </article>)}
  </section>;
}
