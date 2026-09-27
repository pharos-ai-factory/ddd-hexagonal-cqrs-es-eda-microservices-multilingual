import type {Metadata} from 'next';
import './style.css';
export const metadata: Metadata = {title: 'Café Lab', description: 'A small café for exploring dependable services'};
export default function Layout({children}: Readonly<{children: React.ReactNode}>) {
  return <html lang="en-GB"><body>{children}</body></html>;
}
