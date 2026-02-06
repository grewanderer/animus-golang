import type { Metadata, Viewport } from 'next';
import { JetBrains_Mono, Space_Grotesk } from 'next/font/google';

import './globals.css';

const mono = JetBrains_Mono({ subsets: ['latin', 'cyrillic'], variable: '--font-mono' });
const sans = Space_Grotesk({
  subsets: ['latin'],
  variable: '--font-sans',
  fallback: ['Inter', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'sans-serif'],
});

const metadataBase = new URL(
  process.env.NEXT_PUBLIC_SITE_URL ??
    (process.env.NODE_ENV === 'development' ? 'http://localhost:3001' : 'https://console.local'),
);

export const metadata: Metadata = {
  metadataBase,
  title: {
    default: 'Animus Datalab — Консоль управления',
    template: '%s · Animus Datalab',
  },
  description: 'Внутренняя консоль управления контрольной плоскостью Animus Datalab.',
};

export const viewport: Viewport = {
  themeColor: '#081018',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru" data-surface="ops" className={`${sans.variable} ${mono.variable}`}>
      <body data-surface="ops">{children}</body>
    </html>
  );
}
