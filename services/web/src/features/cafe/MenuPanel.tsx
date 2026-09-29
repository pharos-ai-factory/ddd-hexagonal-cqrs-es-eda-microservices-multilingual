'use client';
import {useState} from 'react';
import {items, type Snapshot} from '../realtime/model';
import {minorUnits, money, short, type SendCommand} from './format';

export function MenuPanel({snapshot, send, disabled}: {snapshot: Snapshot; send: SendCommand; disabled: boolean}) {
  const [drinkName, setDrinkName] = useState('Cappuccino');
  const [editionId, setEditionId] = useState('');
  const [drinkId, setDrinkId] = useState('');
  const [offerCode, setOfferCode] = useState('C1');
  const [price, setPrice] = useState('3.00');
  const drinks = items(snapshot, 'drink');
  const editions = items(snapshot, 'edition');
  const selectedEdition = editions.find(item => item.state.id === editionId) ?? editions[0];
  const selectedDrink = drinks.find(item => item.state.id === drinkId);
  const minor = minorUnits(price);

  return <section className="panel menu"><div className="section-title"><span>01 / STOREFRONT</span><h2>The menu</h2></div>
    <div className="form-row"><label>Drink name<input value={drinkName} onChange={e => setDrinkName(e.target.value)}/></label>
      <button disabled={disabled || !drinkName.trim()} onClick={() => void send('menu/drinks/'+crypto.randomUUID(), {name: drinkName}, 0)}>Create drink</button></div>
    {drinks.map(item => <div className="list-row" key={item.state.id}><div><strong>{item.state.name}</strong><small>Revision {item.state.revision || 'draft'}</small></div>
      {item.state.published ? <span className="badge">Published</span> :
        <button disabled={disabled} onClick={() => void send(`menu/drinks/${item.state.id}/publish`, {}, item.version)}>Publish drink</button>}</div>)}
    <div className="divider"/>
    <div className="form-row"><label>Menu edition<select value={selectedEdition?.state.id ?? ''} onChange={e => setEditionId(e.target.value)}>
      <option value="" disabled>Select an edition</option>{editions.map(item => <option key={item.state.id} value={item.state.id}>{short(item.state.id)} · {item.state.status}</option>)}
    </select></label><button disabled={disabled} onClick={async () => {
      const id = crypto.randomUUID(); if (await send('menu/editions/'+id, {currency: 'EUR'}, 0)) setEditionId(id);
    }}>New edition</button></div>
    {selectedEdition && <>
      {selectedEdition.state.offers.map(offer => <div className="list-row" key={offer.code}><span>{offer.code} · {offer.name}</span><strong>{money(offer.minor, offer.currency)}</strong></div>)}
      {selectedEdition.state.status === 'draft' ? <div className="offer-form"><label>Published drink<select value={drinkId} onChange={e => setDrinkId(e.target.value)}>
        <option value="">Choose a drink</option>{drinks.filter(d => d.state.published).map(d => <option key={d.state.id} value={d.state.id}>{d.state.name}</option>)}</select></label>
        <div className="form-row"><label>Code<input value={offerCode} onChange={e => setOfferCode(e.target.value.toUpperCase())}/></label>
          <label>Price (€)<input inputMode="decimal" value={price} onChange={e => setPrice(e.target.value)}/></label></div>
        <div className="form-row"><button disabled={disabled || !selectedDrink || minor === undefined} onClick={() => {
          if (!selectedDrink || minor === undefined) return;
          void send(`menu/editions/${selectedEdition.state.id}/offers`, {code: offerCode, drinkId,
            drinkRevision: selectedDrink.state.revision, minor}, selectedEdition.version);
        }}>Add offer</button><button disabled={disabled || selectedEdition.state.offers.length === 0}
          onClick={() => void send(`menu/editions/${selectedEdition.state.id}/publish`, {}, selectedEdition.version)}>Publish menu</button></div>
      </div> : <p className="note">This edition is published. Its names and prices are frozen.</p>}
    </>}
  </section>;
}
