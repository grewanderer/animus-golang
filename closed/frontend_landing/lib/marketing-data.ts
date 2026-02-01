import { site } from '@/config/site';
import { type Locale } from './i18n';

type NavItem = {
  label: string;
  href: string;
  children?: { label: string; href: string }[];
};
type HeroMetric = { label: string; value: string; detail: string };

type Dataset = {
  marketingNav: NavItem[];
  heroMetrics: HeroMetric[];
  partnerLogos: string[];
};

const datasets: Record<Locale, Dataset> = {
  en: {
    marketingNav: [
      { label: 'Definition', href: '#why' },
      { label: 'Reproducibility', href: '#reproducibility' },
      { label: 'Execution', href: '#process' },
      { label: 'Architecture', href: '#architecture' },
      { label: 'Security', href: '#security' },
      { label: 'Operations', href: '#outcomes' },
      {
        label: 'Docs',
        href: '/docs',
      },
      { label: 'Repository', href: site.repoUrl },
      { label: 'Contact', href: '#contact' },
    ],
    heroMetrics: [
      {
        label: 'Execution unit',
        value: 'Run',
        detail: 'Defined by DatasetVersion, CodeRef, EnvironmentLock',
      },
      {
        label: 'Deployment models',
        value: 'Single / multi-cluster',
        detail: 'On-prem, private cloud, air-gapped',
      },
      { label: 'Audit', value: 'Append-only', detail: 'Exportable AuditEvent' },
    ],
    partnerLogos: ['Control Plane', 'Data Plane', 'Run', 'AuditEvent'],
  },
  ru: {
    marketingNav: [
      { label: 'Определение', href: '#why' },
      { label: 'Воспроизводимость', href: '#reproducibility' },
      { label: 'Исполнение', href: '#process' },
      { label: 'Архитектура', href: '#architecture' },
      { label: 'Безопасность', href: '#security' },
      { label: 'Эксплуатация', href: '#outcomes' },
      {
        label: 'Документация',
        href: '/docs',
      },
      { label: 'Репозиторий', href: site.repoUrl },
      { label: 'Контакт', href: '#contact' },
    ],
    heroMetrics: [
      {
        label: 'Единица исполнения',
        value: 'Run',
        detail: 'DatasetVersion, CodeRef, EnvironmentLock',
      },
      {
        label: 'Модели развёртывания',
        value: 'Single / multi-cluster',
        detail: 'On-prem, private cloud, air-gapped',
      },
      { label: 'Аудит', value: 'Append-only', detail: 'Exportable AuditEvent' },
    ],
    partnerLogos: ['Control Plane', 'Data Plane', 'Run', 'AuditEvent'],
  },
  es: {
    marketingNav: [
      { label: 'Definición', href: '#why' },
      { label: 'Reproducibilidad', href: '#reproducibility' },
      { label: 'Ejecución', href: '#process' },
      { label: 'Arquitectura', href: '#architecture' },
      { label: 'Seguridad', href: '#security' },
      { label: 'Operaciones', href: '#outcomes' },
      {
        label: 'Documentación',
        href: '/docs',
      },
      { label: 'Repositorio', href: site.repoUrl },
      { label: 'Contacto', href: '#contact' },
    ],
    heroMetrics: [
      {
        label: 'Unidad de ejecución',
        value: 'Run',
        detail: 'DatasetVersion, CodeRef, EnvironmentLock',
      },
      {
        label: 'Modelos de despliegue',
        value: 'Single / multi-cluster',
        detail: 'On-prem, nube privada, air-gapped',
      },
      { label: 'Auditoría', value: 'Append-only', detail: 'Exportable AuditEvent' },
    ],
    partnerLogos: ['Control Plane', 'Data Plane', 'Run', 'AuditEvent'],
  },
};

export function getMarketingData(locale: Locale = 'en'): Dataset {
  return datasets[locale] ?? datasets.en;
}
