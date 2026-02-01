import '@/styles/globals.css';
import { defaultLocale } from '@/lib/i18n';
import { baseMetadata } from '@/lib/seo';

export const metadata = baseMetadata;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang={defaultLocale}>
      <body className="min-h-[100dvh] min-h-screen bg-[#040910] text-white">{children}</body>
    </html>
  );
}
