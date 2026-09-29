'use client';
import {useState} from 'react';
import {items, type Snapshot} from '../realtime/model';
import {money, short, type SendCommand} from './format';

export function OrderingPanel({snapshot, customer, send, disabled}: {
  snapshot: Snapshot; customer: string; send: SendCommand; disabled: boolean;
}) {
  const [quantity, setQuantity] = useState(1);
  const editions = items(snapshot, 'edition');
  const orders = items(snapshot, 'order');
  const tickets = items(snapshot, 'ticket');
  const pickups = items(snapshot, 'pickup');
  const latestMenu = editions.filter(item => item.state.status === 'published').at(-1);

  return <section className="panel orders"><div className="section-title"><span>02 / ORDERING</span><h2>At the counter</h2></div>
    <label>Drinks per line<input type="number" min="1" max="5" value={quantity} onChange={e => setQuantity(Number(e.target.value))}/></label>
    <button className="wide" disabled={disabled || !latestMenu || !customer} onClick={() => {
      if (latestMenu) void send('ordering/orders/'+crypto.randomUUID(), {customerId: customer, editionId: latestMenu.state.id}, 0);
    }}>Start an order</button>
    {!latestMenu && <p className="empty">Publish a menu to open the counter.</p>}
    {orders.filter(order => order.state.customerId === customer).map(order => <article className="card" data-testid="order" key={order.state.id}>
      <div className="card-heading"><strong>Order {short(order.state.id)}</strong><span className="badge">{order.state.status}</span></div>
      {order.state.lines.map(line => <p key={line.id}>{line.quantity} × {line.selection.name} <span>{money(line.selection.minor*line.quantity, order.state.currency)}</span></p>)}
      {order.state.status === 'draft' ? <><div className="offer-buttons">
        {editions.find(e => e.state.id === order.state.editionId)?.state.offers.map(offer => <button className="quiet" disabled={disabled} key={offer.code}
          onClick={() => void send(`ordering/orders/${order.state.id}/lines`, {lineId: crypto.randomUUID(),
            editionId: order.state.editionId, offerCode: offer.code, quantity}, order.version)}>+ {offer.name}</button>)}</div>
        <button className="wide" disabled={disabled || order.state.lines.length === 0}
          onClick={() => void send(`ordering/orders/${order.state.id}/place`, {}, order.version)}>Place order</button></> :
        <small>{pickups.some(p => p.state.orderId === order.state.id && p.state.status === 'collected') ? 'Collected' :
          tickets.some(t => t.state.orderId === order.state.id) ? 'With the preparation team' : 'Waiting for preparation'}</small>}
    </article>)}
  </section>;
}
