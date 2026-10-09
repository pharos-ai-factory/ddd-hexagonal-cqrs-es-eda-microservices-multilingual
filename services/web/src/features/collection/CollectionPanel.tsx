'use client';
import {items, type Snapshot} from '../../shared/realtime/model';
import {short, type SendCommand} from '../../shared/ui/format';
import {useState} from 'react';
/** Supports the single handover of each ready pickup. */
export function CollectionPanel({snapshot, send, disabled}: {snapshot: Snapshot; send: SendCommand; disabled: boolean}) {
  const [codes, setCodes] = useState<Record<string, string>>({});
  const pickups = items(snapshot, 'pickup');
  return <>
    <h3>Collection</h3>
    {pickups.filter(p => p.state.status === 'ready').map(pickup => <article className="card" data-testid="pickup" key={pickup.state.id}>
      <div className="card-heading"><strong>Order {short(pickup.state.orderId)}</strong><span className="badge">Ready</span></div>
      <p>Collection code <b data-testid="collection-code">{pickup.state.code}</b></p>
      <div className="form-row"><input aria-label={'Code for '+short(pickup.state.orderId)} placeholder="Enter collection code"
        value={codes[pickup.state.id] ?? ''} onChange={e => setCodes({...codes, [pickup.state.id]: e.target.value.toUpperCase()})}/>
        <button disabled={disabled} onClick={() => void send('collectOrder', pickup.state.id,
          {code: codes[pickup.state.id] ?? ''}, pickup.version)}>Collect order</button></div>
    </article>)}
  </>;
}
