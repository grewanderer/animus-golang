'use client';

import { FormEvent, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/toast-provider';
import { copyText } from '@/lib/clipboard';
import { CONTACT_EMAIL } from '@/lib/contact';
import { createTranslator, type Locale } from '@/lib/i18n';

type Props = { locale: Locale };

export function ContactForm({ locale }: Props) {
  const [status, setStatus] = useState<'idle' | 'sending' | 'submitted' | 'error'>('idle');
  const toast = useToast();
  const copy: Record<
    Locale,
    {
      namePlaceholder: string;
      companyPlaceholder: string;
      emailPlaceholder: string;
      messagePlaceholder: string;
      submitLabel: string;
      sendingLabel: string;
      submittedLabel: string;
      submittedNote: string;
      errorLabel: string;
      errorNote: string;
      toastSentTitle: string;
      toastSentDescription: string;
      toastErrorTitle: string;
      toastErrorDescription: string;
      copyEmailAction: string;
    }
  > = {
    ru: {
      namePlaceholder: 'Имя',
      companyPlaceholder: 'Организация',
      emailPlaceholder: 'Email',
      messagePlaceholder: 'Опишите Project, модель развёртывания и требуемые интеграции',
      submitLabel: 'Отправить запрос',
      sendingLabel: 'Отправляем...',
      submittedLabel: 'Отправлено',
      submittedNote: 'Запрос зафиксирован.',
      errorLabel: 'Ошибка отправки',
      errorNote: 'Попробуйте еще раз или напишите на email.',
      toastSentTitle: 'Запрос отправлен',
      toastSentDescription: 'Запрос зафиксирован.',
      toastErrorTitle: 'Не удалось отправить',
      toastErrorDescription: 'Попробуйте еще раз или напишите на email.',
      copyEmailAction: 'Скопировать email',
    },
    en: {
      namePlaceholder: 'Name',
      companyPlaceholder: 'Organization',
      emailPlaceholder: 'Email',
      messagePlaceholder: 'Describe Project context, deployment model, and required integrations',
      submitLabel: 'Submit request',
      sendingLabel: 'Submitting...',
      submittedLabel: 'Submitted',
      submittedNote: 'Request recorded.',
      errorLabel: 'Submission failed',
      errorNote: 'Please try again or email us.',
      toastSentTitle: 'Submitted',
      toastSentDescription: 'Request recorded.',
      toastErrorTitle: 'Submission failed',
      toastErrorDescription: 'Please try again or email us.',
      copyEmailAction: 'Copy email',
    },
    es: {
      namePlaceholder: 'Nombre',
      companyPlaceholder: 'Organización',
      emailPlaceholder: 'Email',
      messagePlaceholder: 'Describe el Project, el modelo de despliegue y las integraciones requeridas',
      submitLabel: 'Enviar solicitud',
      sendingLabel: 'Enviando...',
      submittedLabel: 'Enviado',
      submittedNote: 'Solicitud registrada.',
      errorLabel: 'Error al enviar',
      errorNote: 'Inténtalo de nuevo o envíanos un email.',
      toastSentTitle: 'Enviado',
      toastSentDescription: 'Solicitud registrada.',
      toastErrorTitle: 'Error al enviar',
      toastErrorDescription: 'Inténtalo de nuevo o envíanos un email.',
      copyEmailAction: 'Copiar email',
    },
  };
  const t = createTranslator(locale, copy);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (status === 'sending') {
      return;
    }
    setStatus('sending');
    const form = event.currentTarget;
    const formData = new FormData(form);
    const name = String(formData.get('name') ?? '').trim();
    const company = String(formData.get('company') ?? '').trim();
    const email = String(formData.get('email') ?? '').trim();
    const message = String(formData.get('message') ?? '').trim();
    const website = String(formData.get('website') ?? '').trim();

    try {
      const response = await fetch('/api/lead', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          name,
          company,
          email,
          message,
          locale,
          pageUrl: typeof window !== 'undefined' ? window.location.href : undefined,
          website,
        }),
      });

      const data = (await response.json().catch(() => null)) as {
        ok?: boolean;
        error?: string;
      } | null;

      if (!response.ok || !data?.ok) {
        throw new Error(data?.error || `HTTP ${response.status}`);
      }

      setStatus('submitted');
      toast.push({
        intent: 'success',
        title: t('toastSentTitle'),
        description: t('toastSentDescription'),
      });
      form.reset();
    } catch (error) {
      console.error('[marketing] lead submit failed', error);
      setStatus('error');
      toast.push({
        intent: 'error',
        title: t('toastErrorTitle'),
        description: t('toastErrorDescription'),
        action: {
          label: t('copyEmailAction'),
          onSelect: () => {
            void copyText(CONTACT_EMAIL);
          },
        },
      });
    }
  };

  return (
    <form className="space-y-4" onSubmit={handleSubmit}>
      <Input name="name" placeholder={t('namePlaceholder')} aria-label={t('namePlaceholder')} required />
      <Input name="company" placeholder={t('companyPlaceholder')} aria-label={t('companyPlaceholder')} />
      <Input
        name="email"
        type="email"
        placeholder={t('emailPlaceholder')}
        aria-label={t('emailPlaceholder')}
        required
      />
      <input
        name="website"
        tabIndex={-1}
        autoComplete="off"
        aria-hidden="true"
        className="absolute left-[-9999px] top-0 h-px w-px opacity-0"
      />
      <Textarea
        name="message"
        rows={4}
        placeholder={t('messagePlaceholder')}
        aria-label={t('messagePlaceholder')}
        required
      />
      <Button
        type="submit"
        size="lg"
        variant="accent"
        className="w-full"
        disabled={status === 'sending' || status === 'submitted'}
      >
        {status === 'sending'
          ? t('sendingLabel')
          : status === 'submitted'
            ? t('submittedLabel')
            : t('submitLabel')}
      </Button>
      {status === 'submitted' ? (
        <p className="text-center text-sm text-white/75">{t('submittedNote')}</p>
      ) : null}
      {status === 'error' ? (
        <p className="text-center text-sm text-white/75">{t('errorNote')}</p>
      ) : null}
    </form>
  );
}
