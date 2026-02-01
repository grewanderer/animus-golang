import { Button } from '@/components/ui/button';
import { createTranslator, type Locale, localizedPath } from '@/lib/i18n';
import { CorePanelVisual } from '@/sections/landing/core-panel-visual';
import { EmailLink } from '@/sections/landing/email-link';

type HeroCopy = {
  kicker: string;
  title: string;
  headline: string;
  descriptionLines: string[];
  ctaDocs: string;
  ctaTalk: string;
  ctaEmail: string;
  trustAnchors: string[];
  statusLabel: string;
  statusValue: string;
  statusNote: string;
  panelTitle: string;
  panelItems: { label: string; description: string; note: string }[];
  snapshotTitle: string;
  snapshotItems: string[];
  deploymentTitle: string;
  deploymentNote: string;
  visualBrand: string;
  visualControl: string;
  visualRun: string;
};

const heroCopy: Record<Locale, HeroCopy> = {
  en: {
    kicker: 'Digital laboratory',
    title: 'Animus',
    headline: 'Governed execution and explicit context.',
    descriptionLines: [
      'Animus Datalab is a corporate digital laboratory for machine learning that organizes the ML lifecycle as a governed, reproducible system within a single operational contour with common rules of execution, security, and audit.',
      'The Control Plane never executes user code; execution occurs in the Data Plane under policy, isolation, and audit.',
    ],
    ctaDocs: 'Read the documentation',
    ctaTalk: 'Request technical discussion',
    ctaEmail: 'Email',
    trustAnchors: ['On-prem', 'Private cloud', 'Air-gapped', 'RBAC + AuditEvent'],
    statusLabel: 'Execution unit',
    statusValue: 'Run',
    statusNote:
      'Run is the minimal unit of execution and reproducibility, defined by DatasetVersion, CodeRef (commit SHA), EnvironmentLock, parameters, and execution policy.',
    panelTitle: 'Control Plane / Data Plane',
    panelItems: [
      {
        label: 'Control Plane',
        description:
          'Governs metadata, policy enforcement, orchestration, and audit for Project-scoped entities.',
        note: 'Never executes user code.',
      },
      {
        label: 'Data Plane',
        description:
          'Executes user code in isolated container environments with explicit resource limits, network policies, and controlled data and Artifact access.',
        note: 'Containerized execution; Kubernetes baseline.',
      },
    ],
    snapshotTitle: 'Operational snapshot',
    snapshotItems: [
      'Run states: queued, running, succeeded, failed, canceled, unknown.',
      'ModelVersion states: draft, validated, approved, deprecated.',
      'AuditEvent: append-only and exportable.',
    ],
    deploymentTitle: 'Deployment models',
    deploymentNote:
      'Single-cluster and multi-cluster deployments are supported across on-prem, private cloud, and air-gapped environments.',
    visualBrand: 'ANIMUS',
    visualControl: 'Control',
    visualRun: 'Run',
  },
  ru: {
    kicker: 'Цифровая лаборатория',
    title: 'Animus',
    headline: 'Управляемое исполнение и явный контекст.',
    descriptionLines: [
      'Animus Datalab — корпоративная цифровая лаборатория машинного обучения, организующая полный жизненный цикл ML-разработки как управляемую и воспроизводимую систему в едином операционном контуре с общими правилами исполнения, безопасности и аудита.',
      'Control Plane никогда не исполняет пользовательский код; исполнение происходит в Data Plane под политиками, изоляцией и аудитом.',
    ],
    ctaDocs: 'Читать документацию',
    ctaTalk: 'Запросить техническое обсуждение',
    ctaEmail: 'Email',
    trustAnchors: ['On-prem', 'Private cloud', 'Air-gapped', 'RBAC + AuditEvent'],
    statusLabel: 'Единица исполнения',
    statusValue: 'Run',
    statusNote:
      'Run — минимальная единица исполнения и воспроизводимости, определяемая DatasetVersion, CodeRef (commit SHA), EnvironmentLock, параметрами и политикой исполнения.',
    panelTitle: 'Control Plane / Data Plane',
    panelItems: [
      {
        label: 'Control Plane',
        description:
          'Управляет метаданными, политиками, оркестрацией и аудитом для Project-сущностей.',
        note: 'Никогда не исполняет пользовательский код.',
      },
      {
        label: 'Data Plane',
        description:
          'Исполняет пользовательский код в изолированных контейнерных окружениях с явными лимитами ресурсов, сетевыми политиками и контролируемым доступом к данным и артефактам.',
        note: 'Контейнерное исполнение; базовая среда — Kubernetes.',
      },
    ],
    snapshotTitle: 'Операционный срез',
    snapshotItems: [
      'Статусы Run: queued, running, succeeded, failed, canceled, unknown.',
      'Статусы ModelVersion: draft, validated, approved, deprecated.',
      'AuditEvent: append-only и экспортируемый.',
    ],
    deploymentTitle: 'Модели развёртывания',
    deploymentNote:
      'Поддерживаются single-cluster и multi-cluster развёртывания в on-prem, private cloud и air-gapped средах.',
    visualBrand: 'ANIMUS',
    visualControl: 'Control',
    visualRun: 'Run',
  },
  es: {
    kicker: 'Laboratorio digital',
    title: 'Animus',
    headline: 'Ejecución gobernada y contexto explícito.',
    descriptionLines: [
      'Animus Datalab es un laboratorio digital corporativo de ML que organiza el ciclo de vida completo del ML como un sistema gobernado y reproducible dentro de un único contorno operativo con reglas comunes de ejecución, seguridad y auditoría.',
      'El Control Plane nunca ejecuta código de usuario; la ejecución ocurre en el Data Plane bajo políticas, aislamiento y auditoría.',
    ],
    ctaDocs: 'Leer la documentación',
    ctaTalk: 'Solicitar discusión técnica',
    ctaEmail: 'Email',
    trustAnchors: ['On-prem', 'Private cloud', 'Air-gapped', 'RBAC + AuditEvent'],
    statusLabel: 'Unidad de ejecución',
    statusValue: 'Run',
    statusNote:
      'Run es la unidad mínima de ejecución y reproducibilidad, definida por DatasetVersion, CodeRef (commit SHA), EnvironmentLock, parámetros y política de ejecución.',
    panelTitle: 'Control Plane / Data Plane',
    panelItems: [
      {
        label: 'Control Plane',
        description:
          'Gobierna metadatos, aplicación de políticas, orquestación y auditoría para entidades con ámbito de Project.',
        note: 'Nunca ejecuta código de usuario.',
      },
      {
        label: 'Data Plane',
        description:
          'Ejecuta código de usuario en entornos de contenedores aislados con límites explícitos de recursos, políticas de red y acceso controlado a datos y Artifacts.',
        note: 'Ejecución en contenedores; línea base: Kubernetes.',
      },
    ],
    snapshotTitle: 'Resumen operativo',
    snapshotItems: [
      'Estados de Run: queued, running, succeeded, failed, canceled, unknown.',
      'Estados de ModelVersion: draft, validated, approved, deprecated.',
      'AuditEvent: append-only y exportable.',
    ],
    deploymentTitle: 'Modelos de despliegue',
    deploymentNote:
      'Se admiten despliegues single-cluster y multi-cluster en entornos on-prem, nube privada y air-gapped.',
    visualBrand: 'ANIMUS',
    visualControl: 'Control',
    visualRun: 'Run',
  },
};

type Props = {
  locale: Locale;
};

export function MarketingHero({ locale }: Props) {
  const t = createTranslator(locale, heroCopy);
  return (
    <section className="relative overflow-hidden rounded-[36px] border border-white/10 bg-[#0a1422]/85 px-6 py-16 shadow-[0_35px_70px_rgba(3,8,18,0.55)] md:px-14">
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_22%,rgba(120,190,230,0.18),transparent_40%),radial-gradient(circle_at_78%_30%,rgba(90,160,210,0.16),transparent_42%),linear-gradient(180deg,rgba(6,14,24,0.65),rgba(4,10,18,0.92))]" />
        <div className="absolute inset-0 opacity-30 mix-blend-screen bg-[url('data:image/svg+xml,%3Csvg width=%22480%22 height=%22480%22 viewBox=%220 0 480 480%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cpath d=%22M0 0H480V480H0z%22 fill=%22transparent%22/%3E%3Cpath d=%22M0 120H480M0 240H480M0 360H480M120 0V480M240 0V480M360 0V480%22 stroke=%22rgba(255,255,255,0.05)%22 stroke-width=%221%22/%3E%3C/svg%3E')]" />
      </div>

      <div className="relative z-10 grid gap-12 lg:grid-cols-[1.15fr_0.95fr]">
        <div className="space-y-7">
          <div className="flex flex-wrap items-center gap-3 text-[11px] uppercase tracking-[0.38em] text-white/60">
            <span className="h-px w-12 bg-white/35" />
            <span>{t('kicker')}</span>
          </div>
          <h1 className="text-4xl font-semibold leading-tight text-white sm:text-5xl lg:text-[56px] lg:leading-[1.02]">
            {t('title')}
          </h1>
          <p className="text-2xl font-semibold text-white/90 sm:text-3xl">{t('headline')}</p>
          <div className="max-w-2xl space-y-3 text-lg text-white/85">
            {t('descriptionLines').map((line) => (
              <p key={line}>{line}</p>
            ))}
          </div>
          <div className="flex flex-wrap gap-3">
            <Button asChild size="lg" variant="accent">
              <a href={localizedPath(locale, '/docs')}>{t('ctaDocs')}</a>
            </Button>
            <Button asChild size="lg" variant="outline">
              <a href="#contact">{t('ctaTalk')}</a>
            </Button>
            <Button asChild size="lg" variant="ghost" className="border border-white/10">
              <EmailLink locale={locale}>{t('ctaEmail')}</EmailLink>
            </Button>
          </div>
          <div className="flex flex-wrap items-center gap-3 text-xs uppercase tracking-[0.22em] text-white/60">
            {t('trustAnchors').map((anchor, index) => (
              <span key={anchor} className="flex items-center gap-3">
                {index > 0 ? <span className="h-1 w-1 rounded-full bg-white/30" /> : null}
                <span>{anchor}</span>
              </span>
            ))}
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div className="rounded-2xl border border-white/15 bg-[#0b1626]/80 p-4 shadow-[0_18px_40px_rgba(5,12,24,0.4)]">
              <div className="flex items-center justify-between text-[11px] uppercase tracking-[0.32em] text-white/60">
                <span>{t('statusLabel')}</span>
                <span className="text-white">{t('statusValue')}</span>
              </div>
              <p className="mt-3 text-sm text-white/80">{t('statusNote')}</p>
            </div>
            <div className="rounded-2xl border border-white/15 bg-[#0b1626]/80 p-4 shadow-[0_18px_40px_rgba(5,12,24,0.35)]">
              <p className="text-[11px] uppercase tracking-[0.32em] text-white/60">
                {t('snapshotTitle')}
              </p>
              <div className="mt-3 space-y-2 text-sm text-white/80">
                {t('snapshotItems').map((item) => (
                  <div key={item} className="flex items-center gap-3">
                    <span className="size-1.5 rounded-full bg-white/40" />
                    <span>{item}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-6">
          <div className="relative overflow-hidden rounded-[28px] border border-white/12 bg-[#0b1626]/90 shadow-[0_28px_60px_rgba(4,10,20,0.6)]">
            <div className="pointer-events-none absolute inset-0">
              <CorePanelVisual className="absolute inset-0 opacity-80" />
              <div className="absolute inset-0 opacity-45 mix-blend-screen bg-[url('data:image/svg+xml,%3Csvg width=%22560%22 height=%22560%22 viewBox=%220 0 560 560%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Crect width=%22560%22 height=%22560%22 fill=%22transparent%22/%3E%3Cg stroke=%22rgba(170,220,255,0.16)%22 stroke-width=%221%22%3E%3Cpath d=%22M140 0 L280 80.8 L280 242.4 L140 323.2 L0 242.4 L0 80.8 Z%22/%3E%3Cpath d=%22M420 0 L560 80.8 L560 242.4 L420 323.2 L280 242.4 L280 80.8 Z%22/%3E%3Cpath d=%22M280 242.4 L420 323.2 L420 484.8 L280 565.6 L140 484.8 L140 323.2 Z%22/%3E%3Cpath d=%22M0 242.4 L140 323.2 L140 484.8 L0 565.6 L-140 484.8 L-140 323.2 Z%22/%3E%3C/g%3E%3C/svg%3E')] bg-[length:560px_560px]" />
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_18%,rgba(120,200,230,0.28),transparent_45%),radial-gradient(circle_at_82%_32%,rgba(90,170,220,0.22),transparent_52%)]" />
            </div>
            <div className="relative z-10 flex min-h-[260px] items-end justify-end px-6 pb-8 pt-10 sm:min-h-[320px]">
              <div className="absolute left-6 top-6 z-10 flex items-center gap-3 text-[10px] uppercase tracking-[0.4em] text-white/60">
                <span className="h-px w-8 bg-white/35" />
                <span>{t('visualBrand')}</span>
                <span className="h-px w-8 bg-white/35" />
              </div>
              <div className="absolute bottom-0 left-6 right-6 z-0 h-24 -skew-x-12 bg-[linear-gradient(90deg,rgba(80,160,210,0.1)_1px,transparent_1px),linear-gradient(0deg,rgba(80,160,210,0.1)_1px,transparent_1px)] bg-[size:28px_28px] opacity-35" />
              <div className="relative z-10 w-full max-w-[360px]">
                <div className="absolute -left-10 -top-10 h-40 w-full rounded-2xl border border-white/10 bg-[#0c1b2c]/65 shadow-[0_14px_28px_rgba(2,8,16,0.45)]" />
                <div className="absolute -left-5 -top-5 h-40 w-full rounded-2xl border border-white/15 bg-[#0b1b2c]/80 shadow-[0_16px_32px_rgba(3,9,18,0.55)]" />
                <div className="relative h-40 w-full rounded-2xl border border-white/20 bg-gradient-to-br from-[#10263b]/95 via-[#0c1d2e]/94 to-[#0a1522]/95 p-4 shadow-[0_20px_40px_rgba(4,12,22,0.65)]">
                  <div className="flex items-center justify-between text-[10px] uppercase tracking-[0.3em] text-white/55">
                    <span>{t('visualControl')}</span>
                    <span className="text-white/75">{t('visualRun')}</span>
                  </div>
                  <div className="mt-3 space-y-2">
                    <div className="h-2 w-4/5 rounded bg-white/25" />
                    <div className="h-2 w-3/4 rounded bg-white/20" />
                    <div className="h-2 w-2/3 rounded bg-white/15" />
                  </div>
                  <div className="mt-4 grid grid-cols-6 gap-2">
                    <div className="col-span-2 h-6 rounded border border-white/15 bg-white/5" />
                    <div className="col-span-4 h-6 rounded border border-white/15 bg-white/5" />
                  </div>
                  <div className="relative mt-3 h-10 overflow-hidden rounded border border-white/15 bg-[#07121d]/80">
                    <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(90,180,220,0.1)_0%,rgba(110,200,240,0.35)_45%,rgba(90,180,220,0.1)_100%)]" />
                    <div className="absolute inset-0 opacity-60 bg-[linear-gradient(90deg,transparent_0%,transparent_12%,rgba(255,255,255,0.18)_14%,transparent_16%,transparent_38%,rgba(255,255,255,0.18)_40%,transparent_42%,transparent_64%,rgba(255,255,255,0.18)_66%,transparent_68%,transparent_100%)]" />
                  </div>
                </div>
                <div className="absolute -right-6 bottom-2 hidden h-24 w-24 rounded-2xl border border-white/20 bg-[#0b1728]/85 p-3 text-white/60 shadow-[0_18px_36px_rgba(3,10,18,0.6)] sm:block">
                  <svg viewBox="0 0 120 120" className="h-full w-full" fill="none" aria-hidden="true">
                    <path
                      d="M18 84 L54 54 L94 70"
                      stroke="rgba(120,200,240,0.55)"
                      strokeWidth="2"
                    />
                    <path
                      d="M28 30 L54 54 L86 34"
                      stroke="rgba(120,200,240,0.35)"
                      strokeWidth="1.6"
                    />
                    <circle cx="18" cy="84" r="5" fill="rgba(160,220,250,0.85)" />
                    <circle cx="54" cy="54" r="6" fill="rgba(200,240,255,0.95)" />
                    <circle cx="94" cy="70" r="5" fill="rgba(140,210,245,0.8)" />
                    <circle cx="28" cy="30" r="4" fill="rgba(160,220,250,0.7)" />
                    <circle cx="86" cy="34" r="4" fill="rgba(160,220,250,0.7)" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="rounded-[28px] border border-white/12 bg-[#0b1626]/90 p-6 shadow-[0_28px_60px_rgba(4,10,20,0.6)]">
            <div className="space-y-6">
              <div className="text-xs uppercase tracking-[0.3em] text-white/60">
                {t('panelTitle')}
              </div>
              <div className="space-y-4">
                {t('panelItems').map((item) => (
                  <div
                    key={item.label}
                    className="rounded-2xl border border-white/12 bg-[#0a1422]/80 p-4"
                  >
                    <div className="text-sm font-semibold text-white">{item.label}</div>
                  <p className="mt-2 text-sm text-white/80">{item.description}</p>
                  <p className="mt-2 text-xs uppercase tracking-[0.24em] text-white/45">
                    {item.note}
                  </p>
                  </div>
                ))}
              </div>
              <div className="rounded-2xl border border-white/12 bg-[#0a1422]/80 p-4 text-sm text-white/75">
                <div className="text-[11px] uppercase tracking-[0.32em] text-white/60">
                  {t('deploymentTitle')}
                </div>
                <p className="mt-2">{t('deploymentNote')}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
