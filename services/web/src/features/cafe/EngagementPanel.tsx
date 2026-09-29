'use client';
import {items, type Snapshot} from '../realtime/model';
import {type SendCommand} from './format';

export function EngagementPanel({snapshot, customer, send, disabled}: {
  snapshot: Snapshot; customer: string; send: SendCommand; disabled: boolean;
}) {
  const accounts = items(snapshot, 'account');
  const rewards = items(snapshot, 'reward');
  const notices = items(snapshot, 'notification');
  const orders = items(snapshot, 'order');
  const customerRewards = rewards.filter(reward => reward.state.customerId === customer);
  const latestOrder = orders.filter(order => order.state.customerId === customer).at(-1);

  return <section className="engagement"><div><div className="section-title"><span>04 / ENGAGEMENT</span><h2>A little thank you.</h2></div>
    {accounts.filter(a => a.state.id === customer).map(account => <div key={account.state.id} data-testid="loyalty-account">
      <div className="stamps">{[0, 1, 2].map(i => <span className={i < account.state.stampBalance ? 'filled' : ''} key={i}>◒</span>)}</div>
      <p>{account.state.collections} orders collected · {account.state.grantsEarned} rewards earned</p>
      {account.state.grantsEarned > customerRewards.length &&
        <p>A grant is earned. Reward issuance is pending.</p>}
    </div>)}
    {customerRewards.map(reward => <article className="reward" data-testid="reward" key={reward.state.id}>
      <strong>{reward.state.benefit}</strong><span className="badge">{reward.state.status}</span>
      <small>Valid until {new Date(reward.state.expiresAt).toLocaleDateString('en-GB')}</small>
      {reward.state.status === 'issued' && <button className="quiet" disabled={disabled || !latestOrder}
        onClick={() => {
          if (latestOrder) void send(`loyalty/rewards/${reward.state.id}/redeem`, {orderId: latestOrder.state.id}, reward.version);
        }}>Record redemption</button>}
    </article>)}</div>
    <div><h3>Customer messages</h3>{notices.filter(n => n.state.recipient === customer).map(notice => <article className="message" data-testid="notification" key={notice.state.id}>
      <div className="card-heading"><strong>{notice.state.subject}</strong><span className="badge">{notice.state.status}</span></div><p>{notice.state.body}</p></article>)}
      <p className="note">Messages go to the local mailbox simulator. No email is sent.</p></div>
  </section>;
}
