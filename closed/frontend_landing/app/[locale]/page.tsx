import type { Metadata } from 'next';
import { notFound } from 'next/navigation';

import { MarketingHero } from '@/sections/landing/hero';
import { ContactForm } from '@/sections/landing/contact-form';
import { EmailLink } from '@/sections/landing/email-link';
import { MarketingSection } from '@/sections/landing/section';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { CONTACT_EMAIL } from '@/lib/contact';
import {
  createTranslator,
  defaultLocale,
  localizedPath,
  resolveLocaleParam,
  type Locale,
} from '@/lib/i18n';
import { buildPageMetadata } from '@/lib/seo';

type PageProps = { params?: Promise<{ locale?: string | string[] }> };

const metaCopy: Record<Locale, { title: string; description: string }> = {
  en: {
    title: 'Animus',
    description:
      'Corporate digital laboratory for machine learning with explicit domain entities, governed Run execution, and Control Plane / Data Plane separation.',
  },
  ru: {
    title: 'Animus',
    description:
      'Корпоративная цифровая лаборатория машинного обучения с явными сущностями, управляемым исполнением Run и разделением Control Plane / Data Plane.',
  },
  es: {
    title: 'Animus',
    description:
      'Laboratorio digital corporativo de ML con entidades explícitas, ejecución gobernada de Run y separación Control Plane / Data Plane.',
  },
};

function getLocaleOrThrow(value: string | string[] | undefined): Locale {
  if (!value) return defaultLocale;
  const resolved = resolveLocaleParam(value);
  if (!resolved) {
    notFound();
  }
  return resolved;
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const resolvedParams = (await params) ?? {};
  const locale = getLocaleOrThrow(resolvedParams.locale);
  const meta = metaCopy[locale] ?? metaCopy.en;
  return buildPageMetadata({ ...meta, path: '/', locale });
}

type Feature = { title: string; description: string };
type Step = { title: string; description: string };
type DocsLink = { label: string; href: string };

type Copy = {
  docsGravityTitle: string;
  docsGravityLinks: DocsLink[];
  whatYouGetEyebrow: string;
  whatYouGetTitle: string;
  whatYouGetSubtitle: string;
  whatYouGetItems: string[];
  whyEyebrow: string;
  whyTitle: string;
  whySubtitle: string;
  whatIsTitle: string;
  whatIsDescription: string;
  whatIsItems: string[];
  whatNotTitle: string;
  whatNotDescription: string;
  whatNotItems: string[];
  whyCtas: DocsLink[];
  capabilitiesEyebrow: string;
  capabilitiesTitle: string;
  capabilitiesSubtitle: string;
  capabilities: Feature[];
  processEyebrow: string;
  processTitle: string;
  processSubtitle: string;
  processStepLabel: (step: number) => string;
  processSteps: Step[];
  architectureEyebrow: string;
  architectureTitle: string;
  architectureSubtitle: string;
  architectureItems: string[];
  architectureCta: string;
  architectureAlt: string;
  architectureSnapshotTitle: string;
  architectureSnapshotDescription: string;
  securityEyebrow: string;
  securityTitle: string;
  securitySubtitle: string;
  securityItems: string[];
  securityCta: string;
  outcomesEyebrow: string;
  outcomesTitle: string;
  outcomesSubtitle: string;
  outcomesScopeTitle: string;
  outcomesScopeDescription: string;
  outcomesScopeItems: string[];
  outcomesDeliverablesTitle: string;
  outcomesDeliverablesDescription: string;
  outcomesDeliverables: string[];
  outcomesNotTitle: string;
  outcomesNotDescription: string;
  outcomesNotItems: string[];
  outcomesCta: string;
  maturityTitle: string;
  maturityNote: string;
  maturityBody: string;
  contactEyebrow: string;
  contactTitle: string;
  contactSubtitle: string;
  contactBullets: string[];
  contactEmailLabel: string;
  contactNextTitle: string;
  contactNextDescription: string;
};

const copy: Record<Locale, Copy> = {
  en: {
    docsGravityTitle: 'Documentation',
    docsGravityLinks: [
      { label: 'Overview', href: '/docs/overview' },
      { label: 'System Definition', href: '/docs/system-definition' },
      { label: 'Domain Model', href: '/docs/domain-model' },
      { label: 'Architecture', href: '/docs/architecture' },
      { label: 'Execution Model', href: '/docs/execution-model' },
      { label: 'Security', href: '/docs/security' },
      { label: 'Operations', href: '/docs/operations' },
    ],
    whatYouGetEyebrow: 'System Definition',
    whatYouGetTitle: 'Architectural invariants',
    whatYouGetSubtitle:
      'The documentation defines these properties as mandatory and non-negotiable.',
    whatYouGetItems: [
      'The Control Plane never executes user code.',
      'A production-run is defined by DatasetVersion, CodeRef (commit SHA), and EnvironmentLock.',
      'All significant actions produce AuditEvent records.',
      'Data, code, environments, and results are explicit, versioned entities.',
      'Hidden state that affects execution is disallowed; results must be explainable from explicit entities.',
    ],
    whyEyebrow: 'System boundaries',
    whyTitle: 'What Animus is / is not',
    whySubtitle:
      'Animus Datalab is a corporate digital laboratory for machine learning that organizes the full ML development lifecycle (data, experiments, training, evaluation, and preparation for industrial use) within a single operational contour governed by common rules of execution, security, and audit.',
    whatIsTitle: 'What Animus is',
    whatIsDescription: 'Explicit domain entities and policy-governed execution.',
    whatIsItems: [
      'A Control Plane that governs metadata, policy enforcement, orchestration, and audit for Project-scoped entities.',
      'A Data Plane that executes user code in isolated environments with explicit resource limits.',
      'Run is the unit of execution and reproducibility; PipelineRun is a DAG of Runs.',
      'Developer Environment provides managed IDE sessions; interactive work is not a Run and is not a production-run.',
    ],
    whatNotTitle: 'What Animus is not',
    whatNotDescription: 'Documented non-goals and boundaries.',
    whatNotItems: [
      'Not a source control system; CodeRef points to external SCM.',
      'Not an IDE or code editor as a product; IDE sessions are managed tools within Developer Environment.',
      'Not a full inference platform.',
    ],
    whyCtas: [
      { label: 'System Definition', href: '/docs/system-definition' },
      { label: 'Architecture', href: '/docs/architecture' },
    ],
    capabilitiesEyebrow: 'Reproducibility',
    capabilitiesTitle: 'Reproducibility contract for Run',
    capabilitiesSubtitle:
      'Reproducibility depends on explicit, immutable inputs and a recorded determinism model.',
    capabilities: [
      {
        title: 'Run (definition)',
        description:
          'Minimal unit of execution and reproducibility that yields Artifacts, execution trace, and AuditEvent.',
      },
      {
        title: 'DatasetVersion',
        description:
          'Runs reference immutable DatasetVersion inputs; data changes require a new DatasetVersion.',
      },
      {
        title: 'CodeRef (commit SHA)',
        description:
          'Production-run requires CodeRef with commit SHA; branches and tags are not permitted.',
      },
      {
        title: 'EnvironmentLock',
        description:
          'Execution uses immutable EnvironmentLock with image digest and dependency checksums.',
      },
      {
        title: 'Parameters + execution policy',
        description:
          'Parameters and execution policy are explicit inputs recorded by Control Plane and applied when forming the Execution Plan.',
      },
      {
        title: 'Determinism model',
        description:
          'Strong and weak reproducibility are distinguished; non-strict cases are explicitly recorded.',
      },
    ],
    processEyebrow: 'Execution model',
    processTitle: 'Declarative execution and pipeline semantics',
    processSubtitle:
      'Control Plane validates, plans, and audits; Data Plane executes isolated workloads.',
    processStepLabel: (step) => `Step ${step}`,
    processSteps: [
      {
        title: 'Declare Run or PipelineRun',
        description:
          'Execution is described declaratively; pipeline specifications define DAG steps and dependencies.',
      },
      {
        title: 'Authorize and validate references',
        description:
          'Control Plane enforces RBAC, validates references, applies policies, and records AuditEvent.',
      },
      {
        title: 'Build Execution Plan',
        description:
          'The plan captures image digest, resources, network policies, and secret references for Data Plane.',
      },
      {
        title: 'Execute in Data Plane',
        description:
          'User code runs in isolated containers; Control Plane never executes user code.',
      },
      {
        title: 'Observe without secret leakage',
        description:
          'Logs, metrics, and traces are collected; secrets must not appear in UI, logs, or Artifacts.',
      },
      {
        title: 'Retry, rerun, replay',
        description:
          'Retries, reruns, and replays create new Runs linked to the original; replay uses the saved Execution Plan.',
      },
    ],
    architectureEyebrow: 'Architecture',
    architectureTitle: 'Control Plane and Data Plane',
    architectureSubtitle:
      'The Control Plane never executes user code; it governs policy, metadata, orchestration, and audit. The Data Plane executes untrusted code in isolated environments.',
    architectureItems: [
      'Trust boundaries distinguish user clients, Control Plane, Data Plane, and external systems.',
      'Control Plane stores metadata and audit as the source of truth and remains consistent during Data Plane failures.',
      'Data Plane executes Runs in containerized environments with explicit resource limits and network policies.',
      'AuditEvent is append-only and exportable, covering administrative actions and execution status changes.',
    ],
    architectureCta: 'Architecture docs',
    architectureAlt: 'Control plane and data plane diagram',
    architectureSnapshotTitle: 'Trust boundaries',
    architectureSnapshotDescription: 'Explicit separation of management and execution.',
    securityEyebrow: 'Security model',
    securityTitle: 'Architectural security model',
    securitySubtitle:
      'Authorization, policy enforcement, and audit are enforced by the platform and are not optional.',
    securityItems: [
      'SSO via OIDC/SAML or local accounts for air-gapped installations; session TTL and audit for logins.',
      'Project-centric RBAC with default deny and object-level enforcement; decisions are audited.',
      'Secrets are provided temporarily via external secret stores and must not appear in UI, logs, metrics, or Artifacts.',
      'Network egress is deny-by-default; external connections are explicitly permitted by policy and audited.',
      'AuditEvent is append-only, non-disableable, and exportable to SIEM/monitoring systems.',
    ],
    securityCta: 'Security docs',
    outcomesEyebrow: 'Operations',
    outcomesTitle: 'Operational readiness and acceptance',
    outcomesSubtitle: 'Deployment, upgrades, and recovery are defined as operational contracts.',
    outcomesScopeTitle: 'Deployment models',
    outcomesScopeDescription: 'Supported topologies and isolation modes.',
    outcomesScopeItems: [
      'Single-cluster deployments (Control Plane + Data Plane).',
      'Multi-cluster deployments with one Control Plane and multiple Data Planes.',
      'On-prem, private cloud, and air-gapped environments.',
    ],
    outcomesDeliverablesTitle: 'Lifecycle operations',
    outcomesDeliverablesDescription: 'Installation, upgrades, and recovery are explicit procedures.',
    outcomesDeliverables: [
      'Helm charts and/or Kustomize manifests with versioned container images.',
      'Controlled upgrades with rollback and schema migrations.',
      'Backup & DR for metadata and audit with defined RPO/RTO.',
    ],
    outcomesNotTitle: 'Failure model',
    outcomesNotDescription: 'Expected degradation behavior is defined and observable.',
    outcomesNotItems: [
      'Control Plane operations are idempotent where possible.',
      'Data Plane failure does not corrupt metadata or audit.',
      'Runs enter diagnostic states (unknown/reconciling) during loss of Data Plane.',
    ],
    outcomesCta: 'Contact',
    maturityTitle: 'Acceptance criteria',
    maturityNote: 'Production-grade definition',
    maturityBody:
      'Animus Datalab is production-grade when a full ML lifecycle is executable within one Project, production-run reproducibility is explicit (or limitations are recorded), audit is end-to-end and exportable, security and access policies enforce permissions, deployments are installable, upgradable, rollback-safe, and no hidden state affects results.',
    contactEyebrow: 'Contact',
    contactTitle: 'Request a technical review',
    contactSubtitle:
      'Use the form to share deployment context, security requirements, and integration constraints.',
    contactBullets: [
      'Specify the intended deployment model (single-cluster, multi-cluster, air-gapped).',
      'List required external systems: database, object storage, IdP, secret store, SIEM.',
      'Identify the Run inputs to be governed: DatasetVersion, CodeRef, EnvironmentLock, parameters, execution policy.',
    ],
    contactEmailLabel: 'Or email',
    contactNextTitle: 'Next steps',
    contactNextDescription:
      'Architecture, security, and operations alignment based on the documentation.',
  },
  ru: {
    docsGravityTitle: 'Документация',
    docsGravityLinks: [
      { label: 'Обзор', href: '/docs/overview' },
      { label: 'Определение системы', href: '/docs/system-definition' },
      { label: 'Доменная модель', href: '/docs/domain-model' },
      { label: 'Архитектура', href: '/docs/architecture' },
      { label: 'Модель исполнения', href: '/docs/execution-model' },
      { label: 'Безопасность', href: '/docs/security' },
      { label: 'Эксплуатация', href: '/docs/operations' },
    ],
    whatYouGetEyebrow: 'Определение системы',
    whatYouGetTitle: 'Архитектурные инварианты',
    whatYouGetSubtitle: 'Документация фиксирует эти свойства как обязательные и ненарушаемые.',
    whatYouGetItems: [
      'Control Plane никогда не исполняет пользовательский код.',
      'Production-run определяется DatasetVersion, CodeRef (commit SHA) и EnvironmentLock.',
      'Все значимые действия порождают AuditEvent.',
      'Данные, код, окружения и результаты представлены как явные версионируемые сущности.',
      'Скрытое состояние, влияющее на исполнение, запрещено; результат должен быть объясним через явные сущности.',
    ],
    whyEyebrow: 'Границы системы',
    whyTitle: 'Что такое Animus / чем он не является',
    whySubtitle:
      'Animus Datalab — корпоративная цифровая лаборатория машинного обучения, организующая полный жизненный цикл ML-разработки (данные, эксперименты, обучение, оценка, подготовка к промышленному использованию) в едином операционном контуре с общими правилами исполнения, безопасности и аудита.',
    whatIsTitle: 'Что такое Animus',
    whatIsDescription: 'Явные доменные сущности и исполнение, управляемое политиками.',
    whatIsItems: [
      'Control Plane управляет метаданными, политиками, оркестрацией и аудитом Project-сущностей.',
      'Data Plane исполняет пользовательский код в изолированных окружениях с явными лимитами ресурсов.',
      'Run — единица исполнения и воспроизводимости; PipelineRun — DAG из Run.',
      'Developer Environment предоставляет управляемые IDE-сессии; интерактивная работа не является Run и не является production-run.',
    ],
    whatNotTitle: 'Чем Animus не является',
    whatNotDescription: 'Задокументированные нецели и границы.',
    whatNotItems: [
      'Не система контроля версий; CodeRef указывает на внешний SCM.',
      'Не IDE и не редактор кода как продукт; IDE-сессии — управляемый инструмент Developer Environment.',
      'Не полноценная inference-платформа.',
    ],
    whyCtas: [
      { label: 'Определение системы', href: '/docs/system-definition' },
      { label: 'Архитектура', href: '/docs/architecture' },
    ],
    capabilitiesEyebrow: 'Воспроизводимость',
    capabilitiesTitle: 'Контракт воспроизводимости Run',
    capabilitiesSubtitle:
      'Воспроизводимость опирается на явные неизменяемые входы и фиксируемую модель детерминизма.',
    capabilities: [
      {
        title: 'Run (определение)',
        description:
          'Минимальная единица исполнения и воспроизводимости; порождает Artifacts, execution trace и AuditEvent.',
      },
      {
        title: 'DatasetVersion',
        description:
          'Run ссылается на неизменяемые DatasetVersion; изменение данных оформляется новой версией.',
      },
      {
        title: 'CodeRef (commit SHA)',
        description:
          'Production-run требует CodeRef с commit SHA; ветки и теги не допускаются.',
      },
      {
        title: 'EnvironmentLock',
        description:
          'Исполнение использует неизменяемый EnvironmentLock с image digest и checksums зависимостей.',
      },
      {
        title: 'Параметры и политика исполнения',
        description:
          'Параметры и политика исполнения — явные входы, фиксируемые Control Plane и применяемые при формировании Execution Plan.',
      },
      {
        title: 'Модель детерминизма',
        description:
          'Сильная и слабая воспроизводимость различаются; статус фиксируется явно.',
      },
    ],
    processEyebrow: 'Модель исполнения',
    processTitle: 'Декларативное исполнение и семантика пайплайнов',
    processSubtitle:
      'Control Plane валидирует, планирует и аудитирует; Data Plane исполняет изолированные нагрузки.',
    processStepLabel: (step) => `Шаг ${step}`,
    processSteps: [
      {
        title: 'Декларировать Run или PipelineRun',
        description:
          'Исполнение описывается декларативно; pipeline specification задаёт DAG шагов и зависимостей.',
      },
      {
        title: 'Авторизовать и проверить ссылки',
        description:
          'Control Plane применяет RBAC, проверяет ссылки, политики и фиксирует AuditEvent.',
      },
      {
        title: 'Сформировать Execution Plan',
        description:
          'План фиксирует image digest, ресурсы, сетевые политики и ссылки на секреты для Data Plane.',
      },
      {
        title: 'Исполнить в Data Plane',
        description:
          'Пользовательский код выполняется в изолированных контейнерах; Control Plane не исполняет пользовательский код.',
      },
      {
        title: 'Наблюдаемость без утечек секретов',
        description:
          'Логи, метрики и трейсы собираются; секреты не должны попадать в UI, логи или Artifacts.',
      },
      {
        title: 'Retry, rerun, replay',
        description:
          'Повторы создают новые Run с явной связью с исходным запуском; replay использует сохранённый Execution Plan.',
      },
    ],
    architectureEyebrow: 'Архитектура',
    architectureTitle: 'Control Plane и Data Plane',
    architectureSubtitle:
      'Control Plane никогда не исполняет пользовательский код; он управляет политиками, метаданными, оркестрацией и аудитом. Data Plane исполняет недоверенный код в изолированных окружениях.',
    architectureItems: [
      'Границы доверия различают пользовательских клиентов, Control Plane, Data Plane и внешние системы.',
      'Control Plane хранит метаданные и аудит как источник истины и сохраняет консистентность при сбоях Data Plane.',
      'Data Plane исполняет Run в контейнерных окружениях с явными лимитами ресурсов и сетевыми политиками.',
      'AuditEvent является append-only и экспортируемым; аудит покрывает административные действия и статусы исполнения.',
    ],
    architectureCta: 'Документация по архитектуре',
    architectureAlt: 'Диаграмма control plane и data plane',
    architectureSnapshotTitle: 'Границы доверия',
    architectureSnapshotDescription: 'Явное разделение управления и исполнения.',
    securityEyebrow: 'Модель безопасности',
    securityTitle: 'Архитектурная модель безопасности',
    securitySubtitle:
      'Авторизация, применение политик и аудит являются обязательными элементами платформы.',
    securityItems: [
      'SSO через OIDC/SAML или локальные учетные записи для air-gapped; TTL сессий и аудит входов.',
      'RBAC на уровне Project с default deny и object-level enforcement; решения фиксируются в аудите.',
      'Секреты предоставляются временно через внешние secret store и не должны попадать в UI, логи, метрики или Artifacts.',
      'Сетевой egress по умолчанию запрещён; внешние соединения разрешаются политиками и аудитируются.',
      'AuditEvent append-only, не отключаем и экспортируем в SIEM/monitoring.',
    ],
    securityCta: 'Документация по безопасности',
    outcomesEyebrow: 'Эксплуатация',
    outcomesTitle: 'Эксплуатационная готовность и приёмка',
    outcomesSubtitle:
      'Развёртывание, обновления и восстановление описаны как операционные контракты.',
    outcomesScopeTitle: 'Модели развёртывания',
    outcomesScopeDescription: 'Поддерживаемые топологии и режимы изоляции.',
    outcomesScopeItems: [
      'Single-cluster развёртывания (Control Plane + Data Plane).',
      'Multi-cluster развёртывания с одним Control Plane и несколькими Data Plane.',
      'On-prem, private cloud и air-gapped среды.',
    ],
    outcomesDeliverablesTitle: 'Операционные процедуры',
    outcomesDeliverablesDescription: 'Установка, обновления и восстановление описаны явно.',
    outcomesDeliverables: [
      'Helm charts и/или Kustomize-манифесты с версионированными контейнерными образами.',
      'Контролируемые обновления с rollback и миграциями схем.',
      'Backup & DR для метаданных и аудита с определёнными RPO/RTO.',
    ],
    outcomesNotTitle: 'Модель отказов',
    outcomesNotDescription: 'Ожидаемое поведение при деградации определено и наблюдаемо.',
    outcomesNotItems: [
      'Операции Control Plane идемпотентны там, где это возможно.',
      'Отказ Data Plane не нарушает метаданные и аудит.',
      'Run переходят в диагностические статусы (unknown/reconciling) при потере Data Plane.',
    ],
    outcomesCta: 'Контакт',
    maturityTitle: 'Критерии приёмки',
    maturityNote: 'Определение production-grade',
    maturityBody:
      'Animus Datalab считается production-grade, когда полный ML-цикл выполняется в рамках одного Project, воспроизводимость production-run формализована (или ограничения фиксируются), аудит сквозной и экспортируемый, безопасность и доступы работают end-to-end, развёртывание/обновления/rollback предсказуемы, а скрытое состояние отсутствует.',
    contactEyebrow: 'Контакт',
    contactTitle: 'Запросить технический обзор',
    contactSubtitle:
      'Используйте форму, чтобы передать контекст развёртывания, требования безопасности и ограничения интеграций.',
    contactBullets: [
      'Укажите модель развёртывания (single-cluster, multi-cluster, air-gapped).',
      'Перечислите внешние системы: база данных, объектное хранилище, IdP, secret store, SIEM.',
      'Определите входы Run для управления: DatasetVersion, CodeRef, EnvironmentLock, параметры, execution policy.',
    ],
    contactEmailLabel: 'Или email',
    contactNextTitle: 'Следующие шаги',
    contactNextDescription:
      'Согласование архитектуры, безопасности и эксплуатации на основе документации.',
  },
  es: {
    docsGravityTitle: 'Documentación',
    docsGravityLinks: [
      { label: 'Resumen', href: '/docs/overview' },
      { label: 'Definición del sistema', href: '/docs/system-definition' },
      { label: 'Modelo de dominio', href: '/docs/domain-model' },
      { label: 'Arquitectura', href: '/docs/architecture' },
      { label: 'Modelo de ejecución', href: '/docs/execution-model' },
      { label: 'Seguridad', href: '/docs/security' },
      { label: 'Operaciones', href: '/docs/operations' },
    ],
    whatYouGetEyebrow: 'Definición del sistema',
    whatYouGetTitle: 'Invariantes arquitectónicos',
    whatYouGetSubtitle:
      'La documentación define estas propiedades como obligatorias y no negociables.',
    whatYouGetItems: [
      'El Control Plane nunca ejecuta código de usuario.',
      'Un production-run se define por DatasetVersion, CodeRef (commit SHA) y EnvironmentLock.',
      'Todas las acciones significativas generan AuditEvent.',
      'Datos, código, entornos y resultados son entidades explícitas y versionadas.',
      'El estado oculto que afecta la ejecución está prohibido; los resultados deben ser explicables a partir de entidades explícitas.',
    ],
    whyEyebrow: 'Límites del sistema',
    whyTitle: 'Qué es Animus / qué no es',
    whySubtitle:
      'Animus Datalab es un laboratorio digital corporativo de ML que organiza el ciclo de vida completo del ML (datos, experimentos, entrenamiento, evaluación y preparación para uso industrial) dentro de un único contorno operativo con reglas comunes de ejecución, seguridad y auditoría.',
    whatIsTitle: 'Qué es Animus',
    whatIsDescription: 'Entidades de dominio explícitas y ejecución gobernada por políticas.',
    whatIsItems: [
      'Un Control Plane que gobierna metadatos, aplicación de políticas, orquestación y auditoría para entidades con ámbito de Project.',
      'Un Data Plane que ejecuta código de usuario en entornos aislados con límites explícitos de recursos.',
      'Run es la unidad de ejecución y reproducibilidad; PipelineRun es un DAG de Runs.',
      'Developer Environment proporciona sesiones IDE gestionadas; el trabajo interactivo no es un Run ni un production-run.',
    ],
    whatNotTitle: 'Qué no es Animus',
    whatNotDescription: 'No objetivos y límites documentados.',
    whatNotItems: [
      'No es un sistema de control de versiones; CodeRef apunta a un SCM externo.',
      'No es un IDE ni un editor de código como producto; las sesiones IDE son herramientas gestionadas dentro de Developer Environment.',
      'No es una plataforma completa de inferencia.',
    ],
    whyCtas: [
      { label: 'Definición del sistema', href: '/docs/system-definition' },
      { label: 'Arquitectura', href: '/docs/architecture' },
    ],
    capabilitiesEyebrow: 'Reproducibilidad',
    capabilitiesTitle: 'Contrato de reproducibilidad para Run',
    capabilitiesSubtitle:
      'La reproducibilidad depende de entradas explícitas e inmutables y de un modelo de determinismo registrado.',
    capabilities: [
      {
        title: 'Run (definición)',
        description:
          'Unidad mínima de ejecución y reproducibilidad que produce Artifacts, execution trace y AuditEvent.',
      },
      {
        title: 'DatasetVersion',
        description:
          'Los Runs referencian DatasetVersion inmutables; los cambios de datos requieren una nueva DatasetVersion.',
      },
      {
        title: 'CodeRef (commit SHA)',
        description:
          'Production-run requiere CodeRef con commit SHA; las ramas y etiquetas no están permitidas.',
      },
      {
        title: 'EnvironmentLock',
        description:
          'La ejecución utiliza EnvironmentLock inmutable con image digest y checksums de dependencias.',
      },
      {
        title: 'Parámetros + política de ejecución',
        description:
          'Los parámetros y la política de ejecución son entradas explícitas registradas por el Control Plane y aplicadas al formar el Execution Plan.',
      },
      {
        title: 'Modelo de determinismo',
        description:
          'Se distinguen reproducibilidad fuerte y débil; los casos no estrictos se registran explícitamente.',
      },
    ],
    processEyebrow: 'Modelo de ejecución',
    processTitle: 'Ejecución declarativa y semántica de pipelines',
    processSubtitle:
      'El Control Plane valida, planifica y audita; el Data Plane ejecuta cargas aisladas.',
    processStepLabel: (step) => `Paso ${step}`,
    processSteps: [
      {
        title: 'Declarar Run o PipelineRun',
        description:
          'La ejecución se describe de forma declarativa; las especificaciones de pipeline definen pasos y dependencias DAG.',
      },
      {
        title: 'Autorizar y validar referencias',
        description:
          'El Control Plane aplica RBAC, valida referencias, aplica políticas y registra AuditEvent.',
      },
      {
        title: 'Construir Execution Plan',
        description:
          'El plan captura image digest, recursos, políticas de red y referencias de secretos para el Data Plane.',
      },
      {
        title: 'Ejecutar en Data Plane',
        description:
          'El código de usuario se ejecuta en contenedores aislados; el Control Plane nunca ejecuta código de usuario.',
      },
      {
        title: 'Observabilidad sin filtración de secretos',
        description:
          'Se recopilan logs, métricas y trazas; los secretos no deben aparecer en UI, logs o Artifacts.',
      },
      {
        title: 'Retry, rerun, replay',
        description:
          'Los reintentos, reruns y replays crean nuevos Runs vinculados al original; replay usa el Execution Plan guardado.',
      },
    ],
    architectureEyebrow: 'Arquitectura',
    architectureTitle: 'Control Plane y Data Plane',
    architectureSubtitle:
      'El Control Plane nunca ejecuta código de usuario; gobierna políticas, metadatos, orquestación y auditoría. El Data Plane ejecuta código no confiable en entornos aislados.',
    architectureItems: [
      'Los límites de confianza distinguen clientes de usuario, Control Plane, Data Plane y sistemas externos.',
      'El Control Plane almacena metadatos y auditoría como fuente de verdad y mantiene consistencia durante fallos del Data Plane.',
      'El Data Plane ejecuta Runs en entornos de contenedores con límites explícitos de recursos y políticas de red.',
      'AuditEvent es append-only y exportable, cubriendo acciones administrativas y cambios de estado de ejecución.',
    ],
    architectureCta: 'Documentación de arquitectura',
    architectureAlt: 'Diagrama de control plane y data plane',
    architectureSnapshotTitle: 'Límites de confianza',
    architectureSnapshotDescription: 'Separación explícita de gestión y ejecución.',
    securityEyebrow: 'Modelo de seguridad',
    securityTitle: 'Modelo de seguridad arquitectónico',
    securitySubtitle:
      'La autorización, la aplicación de políticas y la auditoría son obligatorias.',
    securityItems: [
      'SSO via OIDC/SAML o cuentas locales para air-gapped; TTL de sesión y auditoría de inicios de sesión.',
      'RBAC centrado en Project con default deny y enforcement a nivel de objeto; decisiones auditadas.',
      'Los secretos se entregan temporalmente vía secret stores externos y no deben aparecer en UI, logs, métricas o Artifacts.',
      'El egress de red es deny-by-default; conexiones externas permitidas explícitamente por política y auditadas.',
      'AuditEvent es append-only, no desactivable y exportable a sistemas SIEM/monitorización.',
    ],
    securityCta: 'Documentación de seguridad',
    outcomesEyebrow: 'Operaciones',
    outcomesTitle: 'Preparación operativa y aceptación',
    outcomesSubtitle:
      'El despliegue, las actualizaciones y la recuperación se definen como contratos operativos.',
    outcomesScopeTitle: 'Modelos de despliegue',
    outcomesScopeDescription: 'Topologías y modos de aislamiento compatibles.',
    outcomesScopeItems: [
      'Despliegues single-cluster (Control Plane + Data Plane).',
      'Despliegues multi-cluster con un Control Plane y múltiples Data Plane.',
      'Entornos on-prem, nube privada y air-gapped.',
    ],
    outcomesDeliverablesTitle: 'Operaciones del ciclo de vida',
    outcomesDeliverablesDescription:
      'Instalación, actualizaciones y recuperación como procedimientos explícitos.',
    outcomesDeliverables: [
      'Helm charts y/o manifiestos Kustomize con imágenes de contenedor versionadas.',
      'Actualizaciones controladas con rollback y migraciones de esquema.',
      'Backup & DR para metadatos y auditoría con RPO/RTO definidos.',
    ],
    outcomesNotTitle: 'Modelo de fallos',
    outcomesNotDescription: 'El comportamiento esperado en degradación está definido y es observable.',
    outcomesNotItems: [
      'Las operaciones del Control Plane son idempotentes cuando es posible.',
      'El fallo del Data Plane no corrompe metadatos ni auditoría.',
      'Los Runs entran en estados diagnósticos (unknown/reconciling) durante la pérdida del Data Plane.',
    ],
    outcomesCta: 'Contacto',
    maturityTitle: 'Criterios de aceptación',
    maturityNote: 'Definición production-grade',
    maturityBody:
      'Animus Datalab es production-grade cuando un ciclo completo de ML es ejecutable en un Project, la reproducibilidad de production-run es explícita (o se registran sus límites), la auditoría es end-to-end y exportable, la seguridad y los accesos se aplican de extremo a extremo, el despliegue/actualización/rollback es predecible y no existe estado oculto que afecte resultados.',
    contactEyebrow: 'Contacto',
    contactTitle: 'Solicitar revisión técnica',
    contactSubtitle:
      'Utiliza el formulario para compartir contexto de despliegue, requisitos de seguridad y restricciones de integración.',
    contactBullets: [
      'Especifica el modelo de despliegue previsto (single-cluster, multi-cluster, air-gapped).',
      'Enumera sistemas externos requeridos: base de datos, almacenamiento de objetos, IdP, secret store, SIEM.',
      'Identifica las entradas de Run a gobernar: DatasetVersion, CodeRef, EnvironmentLock, parámetros, execution policy.',
    ],
    contactEmailLabel: 'O email',
    contactNextTitle: 'Siguientes pasos',
    contactNextDescription:
      'Alineación de arquitectura, seguridad y operaciones basada en la documentación.',
  },
};

export default async function MarketingPage({ params }: PageProps) {
  const resolvedParams = (await params) ?? {};
  const locale = getLocaleOrThrow(resolvedParams.locale);
  const t = createTranslator(locale, copy);
  return (
    <>
      <MarketingHero locale={locale} />

      <section className="rounded-3xl border border-white/10 bg-[#0b1626]/75 px-6 py-6 shadow-[0_24px_55px_rgba(4,10,20,0.45)]">
        <div className="flex flex-col gap-3">
          <h2 className="text-xl font-semibold text-white">{t('docsGravityTitle')}</h2>
          <div className="flex flex-wrap items-center gap-3 text-sm text-white/80">
            {t('docsGravityLinks').map((item) => (
              <a
                key={item.href}
                href={localizedPath(locale, item.href)}
                className="max-w-full rounded-full border border-white/10 px-3 py-1 text-xs uppercase tracking-[0.2em] text-white/70 hover:border-white/30 hover:text-white break-words leading-tight"
              >
                {item.label}
              </a>
            ))}
          </div>
        </div>
      </section>

      <MarketingSection
        eyebrow={t('whatYouGetEyebrow')}
        title={t('whatYouGetTitle')}
        subtitle={t('whatYouGetSubtitle')}
      >
        <Card className="border-white/12 bg-[#0b1626]/85">
          <CardContent>
            <ul className="space-y-3 text-sm text-white/80">
              {t('whatYouGetItems').map((item) => (
                <li key={item} className="flex items-start gap-3">
                  <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </MarketingSection>

      <MarketingSection
        id="why"
        eyebrow={t('whyEyebrow')}
        title={t('whyTitle')}
        subtitle={t('whySubtitle')}
        actions={
          <div className="flex flex-wrap gap-3">
            {t('whyCtas').map((cta) => (
              <Button key={cta.href} asChild size="sm" variant="outline">
                <a href={localizedPath(locale, cta.href)}>{cta.label}</a>
              </Button>
            ))}
          </div>
        }
      >
        <div className="grid gap-6 lg:grid-cols-2">
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('whatIsTitle')}</CardTitle>
              <CardDescription>{t('whatIsDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('whatIsItems').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('whatNotTitle')}</CardTitle>
              <CardDescription>{t('whatNotDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('whatNotItems').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>
      </MarketingSection>

      <MarketingSection
        id="reproducibility"
        eyebrow={t('capabilitiesEyebrow')}
        title={t('capabilitiesTitle')}
        subtitle={t('capabilitiesSubtitle')}
      >
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {t('capabilities').map((feature) => (
            <Card key={feature.title} className="border-white/12 bg-[#0b1626]/85">
              <CardHeader>
                <CardTitle>{feature.title}</CardTitle>
                <CardDescription>{feature.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </MarketingSection>

      <MarketingSection
        id="process"
        eyebrow={t('processEyebrow')}
        title={t('processTitle')}
        subtitle={t('processSubtitle')}
      >
        <div className="grid gap-4 md:grid-cols-2">
          {t('processSteps').map((step, index) => (
            <Card key={step.title} className="border-white/12 bg-[#0b1626]/85">
              <CardHeader>
                <div className="text-xs uppercase tracking-[0.3em] text-white/60">
                  {t('processStepLabel')(index + 1)}
                </div>
                <CardTitle>{step.title}</CardTitle>
                <CardDescription>{step.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </MarketingSection>

      <MarketingSection
        id="architecture"
        eyebrow={t('architectureEyebrow')}
        title={t('architectureTitle')}
        subtitle={t('architectureSubtitle')}
        actions={
          <Button asChild size="sm" variant="outline">
            <a href={localizedPath(locale, '/docs/architecture')}>{t('architectureCta')}</a>
          </Button>
        }
      >
        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <Card className="border-white/12 bg-[#0b1626]/85 p-6">
            <img
              src="/assets/diagram-control-plane.svg"
              alt={t('architectureAlt')}
              className="w-full rounded-2xl border border-white/10 bg-[#09121e]/80 p-4"
            />
          </Card>
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('architectureSnapshotTitle')}</CardTitle>
              <CardDescription>{t('architectureSnapshotDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('architectureItems').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>
      </MarketingSection>

      <MarketingSection
        id="security"
        eyebrow={t('securityEyebrow')}
        title={t('securityTitle')}
        subtitle={t('securitySubtitle')}
        actions={
          <Button asChild size="sm" variant="outline">
            <a href={localizedPath(locale, '/docs/security')}>{t('securityCta')}</a>
          </Button>
        }
      >
        <Card className="border-white/12 bg-[#0b1626]/85">
          <CardContent>
            <ul className="grid gap-3 text-sm text-white/80 md:grid-cols-2">
              {t('securityItems').map((item) => (
                <li key={item} className="flex items-start gap-3">
                  <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </MarketingSection>

      <MarketingSection
        id="outcomes"
        eyebrow={t('outcomesEyebrow')}
        title={t('outcomesTitle')}
        subtitle={t('outcomesSubtitle')}
        actions={
          <Button asChild size="sm" variant="outline">
            <a href="#contact">{t('outcomesCta')}</a>
          </Button>
        }
      >
        <div className="grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('outcomesScopeTitle')}</CardTitle>
              <CardDescription>{t('outcomesScopeDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('outcomesScopeItems').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('outcomesDeliverablesTitle')}</CardTitle>
              <CardDescription>{t('outcomesDeliverablesDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('outcomesDeliverables').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>
        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('outcomesNotTitle')}</CardTitle>
              <CardDescription>{t('outcomesNotDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('outcomesNotItems').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('maturityTitle')}</CardTitle>
              <CardDescription>{t('maturityNote')}</CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-white/80">
              <p>{t('maturityBody')}</p>
            </CardContent>
          </Card>
        </div>
      </MarketingSection>

      <MarketingSection
        id="contact"
        eyebrow={t('contactEyebrow')}
        title={t('contactTitle')}
        subtitle={t('contactSubtitle')}
      >
        <div className="grid gap-6 lg:grid-cols-[1fr_0.9fr]">
          <Card className="border-white/12 bg-[#0b1626]/85 p-6">
            <ContactForm locale={locale} />
          </Card>
          <Card className="border-white/12 bg-[#0b1626]/85">
            <CardHeader>
              <CardTitle>{t('contactNextTitle')}</CardTitle>
              <CardDescription>{t('contactNextDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3 text-sm text-white/80">
                {t('contactBullets').map((item) => (
                  <li key={item} className="flex items-start gap-3">
                    <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-white/50" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
              <div className="mt-6 flex items-center gap-2 text-sm text-white/75">
                <span>{t('contactEmailLabel')}</span>
                <EmailLink locale={locale} className="text-white underline">
                  {CONTACT_EMAIL}
                </EmailLink>
              </div>
            </CardContent>
          </Card>
        </div>
      </MarketingSection>
    </>
  );
}
