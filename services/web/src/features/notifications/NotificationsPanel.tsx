'use client';
import {items, type Snapshot} from '../../shared/realtime/model';
/** Displays the customer's delivery status from the Communication projection. */
export function NotificationsPanel({snapshot, customer}: {snapshot: Snapshot; customer: string}) {
  const notices = items(snapshot, 'notification');
  return (
    <div><h3>Customer messages</h3>{notices.filter(n => n.state.recipient === customer).map(notice => <article className="message" data-testid="notification" key={notice.state.id}>
      <div className="card-heading"><strong>{notice.state.subject}</strong><span className="badge">{notice.state.status}</span></div><p>{notice.state.body}</p></article>)}
      <p className="note">Messages go to the local mailbox simulator. No email is sent.</p></div>
  );
}
