'use client';

import { createTranslator, defaultLocale, isLocale, locales, type Locale } from '@/lib/i18n';

type GlobalErrorProps = {
  error: Error & { digest?: string };
  reset: () => void;
};

const copy: Record<
  Locale,
  { label: string; title: string; description: string; debugLabel: string; retry: string }
> = {
  en: {
    label: 'Animus · error',
    title: 'We could not render this page',
    description: 'An error occurred while rendering on the server.',
    debugLabel: 'Debug message:',
    retry: 'Try again',
  },
  ru: {
    label: 'Animus · ошибка',
    title: 'Не удалось отрендерить страницу',
    description: 'Ошибка произошла при рендеринге на сервере.',
    debugLabel: 'Отладочное сообщение:',
    retry: 'Попробовать снова',
  },
  es: {
    label: 'Animus · error',
    title: 'No pudimos renderizar esta página',
    description: 'Ocurrió un error al renderizar en el servidor.',
    debugLabel: 'Mensaje de depuración:',
    retry: 'Intentar de nuevo',
  },
};

function resolveLocale(): Locale {
  if (typeof window === 'undefined') return defaultLocale;
  const [, first] = window.location.pathname.split('/');
  if (isLocale(first)) return first;
  if (locales.includes(first as Locale)) return first as Locale;
  return defaultLocale;
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  // Log the original error so it appears in container logs.
  // This should help diagnose server-side render failures in production.
  console.error('[app] global error', error);

  const locale = resolveLocale();
  const t = createTranslator(locale, copy);

  return (
    <html>
      <body className="min-h-screen bg-[#040910] text-white">
        <div className="mx-auto flex max-w-3xl flex-col gap-4 px-6 py-16">
          <p className="text-xs uppercase tracking-[0.3em] text-white/50">{t('label')}</p>
          <h1 className="text-2xl font-semibold">{t('title')}</h1>
          <div className="text-white/75">
            <p>{t('description')}</p>
            <p className="mt-2 text-sm text-white/70">
              <span>{t('debugLabel')}</span>
              <code className="ml-2">{error.message}</code>
              {error.digest ? (
                <span className="ml-2 text-xs text-white/50">({error.digest})</span>
              ) : null}
            </p>
          </div>
          <button
            type="button"
            onClick={reset}
            className="w-fit rounded-lg border border-white/15 bg-white/5 px-4 py-2 text-sm text-white hover:border-white/25"
          >
            {t('retry')}
          </button>
        </div>
      </body>
    </html>
  );
}
