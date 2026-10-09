'use client';
import {items, type Snapshot} from '../../shared/realtime/model';
import {short, type SendCommand} from '../../shared/ui/format';
/** Shows the preparation work queue and its available commands. */
export function PreparationPanel({snapshot, send, disabled}: {snapshot: Snapshot; send: SendCommand; disabled: boolean}) {
  const tickets = items(snapshot, 'ticket');
  return <>
    <h3>Preparation</h3>
    {!tickets.length && <p className="empty">Placed orders arrive here automatically.</p>}
    {tickets.filter(t => t.state.status !== 'ready').map(ticket => <article className="card" data-testid="ticket" key={ticket.state.id}>
      <div className="card-heading"><strong>Order {short(ticket.state.orderId)}</strong><span className="badge">{ticket.state.status}</span></div>
      <p>{ticket.state.instructions}</p><button disabled={disabled} onClick={() => void send(
        ticket.state.status === 'queued' ? 'startPreparation' : 'completePreparation', ticket.state.id, {}, ticket.version)}>
        {ticket.state.status === 'queued' ? 'Start preparation' : 'Mark drinks ready'}</button>
    </article>)}
  </>;
}
