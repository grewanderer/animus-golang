import Link from 'next/link';

import { createTranslator, localizedPath, type Locale } from '@/lib/i18n';
import { cn } from '@/lib/utils';

type Breadcrumb = {
  label: string;
  href?: string;
};

type Props = {
  locale: Locale;
  items: Breadcrumb[];
  className?: string;
};

export function DocsBreadcrumbs({ locale, items, className }: Props) {
  const copy: Record<Locale, { ariaLabel: string }> = {
    en: { ariaLabel: 'Breadcrumb' },
    ru: { ariaLabel: 'Навигация' },
    es: { ariaLabel: 'Miga de pan' },
  };
  const t = createTranslator(locale, copy);
  return (
    <nav aria-label={t('ariaLabel')} className={cn('text-xs text-white/60', className)}>
      <ol className="flex flex-wrap items-center gap-2">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;
          const content = item.href ? (
            <Link href={localizedPath(locale, item.href)} className="hover:text-white">
              {item.label}
            </Link>
          ) : (
            <span className="text-white">{item.label}</span>
          );

          return (
            <li key={`${item.label}-${index}`} className="flex items-center gap-2">
              {content}
              {isLast ? null : <span className="text-white/40">/</span>}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
