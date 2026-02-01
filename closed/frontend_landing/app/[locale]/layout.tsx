import type { ReactNode } from 'react';
import Link from 'next/link';
import { notFound } from 'next/navigation';

import { AppProviders } from '@/app/providers';
import { MarketingNav } from '@/sections/landing/marketing-nav';
import { LiveFrame } from '@/sections/landing/live-frame';
import { BackToTop } from '@/sections/landing/back-to-top';
import { MobileNav } from '@/sections/landing/mobile-nav';
import {
  createTranslator,
  defaultLocale,
  locales,
  localizedPath,
  resolveLocaleParam,
  type Locale,
} from '@/lib/i18n';

type Props = {
  children: ReactNode;
  params?: Promise<{ locale?: string | string[] }>;
};

function getLocaleOrThrow(value: string | string[] | undefined): Locale {
  if (!value) return defaultLocale;
  const resolved = resolveLocaleParam(value);
  if (!resolved) {
    notFound();
  }
  return resolved;
}

export const dynamicParams = false;

export default async function MarketingLayout({ children, params }: Props) {
  const resolvedParams = (await params) ?? {};
  const locale = getLocaleOrThrow(resolvedParams.locale);
  const currentYear = new Date().getUTCFullYear();
  const copy: Record<
    Locale,
    { brand: string; footer: (year: number) => string; localeLabel: string }
  > = {
    en: {
      brand: 'Animus',
      footer: (year) => `© ${year} ANIMUS.`,
      localeLabel: 'Select language',
    },
    ru: {
      brand: 'Animus',
      footer: (year) => `© ${year} ANIMUS.`,
      localeLabel: 'Выбор языка',
    },
    es: {
      brand: 'Animus',
      footer: (year) => `© ${year} ANIMUS.`,
      localeLabel: 'Seleccionar idioma',
    },
  };
  const t = createTranslator(locale, copy);
  return (
    <AppProviders locale={locale}>
      <div
        className="min-h-[100dvh] min-h-screen bg-[#040910] text-white"
        lang={locale}
        data-locale={locale}
      >
        <div className="marketing-shell">
          <div id="top" aria-hidden="true" />
          <LiveFrame />
          <BackToTop locale={locale} />
          <div className="relative z-10 mx-auto flex min-h-[100dvh] min-h-screen w-full max-w-6xl flex-col gap-12 px-4 py-8 text-white sm:px-6 lg:px-10">
            <header className="flex flex-col gap-4 rounded-[32px] border border-white/10 bg-[#0b1626]/85 px-6 py-4 text-sm backdrop-blur-[2px]">
              <div className="flex flex-wrap items-center gap-4">
                <Link
                  href={localizedPath(locale, '#top')}
                  className="font-semibold uppercase tracking-[0.6em] text-white"
                >
                  {t('brand')}
                </Link>
                <div className="ml-auto flex items-center gap-3">
                  <MobileNav locale={locale} />
                  <div
                    className="flex items-center gap-2 rounded-full border border-white/10 px-3 py-1 text-xs uppercase tracking-[0.2em] text-white/70"
                    aria-label={t('localeLabel')}
                  >
                    {locales.map((item) => (
                      <Link
                        key={item}
                        href={localizedPath(item, '/')}
                        className={item === locale ? 'font-semibold text-white' : 'hover:text-white'}
                      >
                        {item}
                      </Link>
                    ))}
                  </div>
                </div>
              </div>
              <MarketingNav locale={locale} />
            </header>

            <main className="flex-1 space-y-16">{children}</main>

            <footer className="rounded-[32px] border border-white/10 bg-[#0b1626]/85 p-6 text-white/75 backdrop-blur-[2px]">
              <p className="text-sm">{t('footer')(currentYear)}</p>
            </footer>
          </div>
        </div>
      </div>
    </AppProviders>
  );
}

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}
