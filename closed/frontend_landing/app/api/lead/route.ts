import { NextResponse } from 'next/server';

export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';

type LeadPayload = {
  name?: unknown;
  company?: unknown;
  email?: unknown;
  message?: unknown;
  locale?: unknown;
  pageUrl?: unknown;
  website?: unknown;
};

function asTrimmedString(value: unknown) {
  if (typeof value !== 'string') return '';
  return value.trim();
}

function truncate(value: string, max: number) {
  if (value.length <= max) return value;
  return `${value.slice(0, Math.max(0, max - 1))}…`;
}

function formatLeadMessage(
  payload: Required<Pick<LeadPayload, 'name' | 'email' | 'message'>> & Partial<LeadPayload>,
  meta: {
    receivedAtIso: string;
    ip?: string | null;
    userAgent?: string | null;
  },
) {
  const name = asTrimmedString(payload.name);
  const company = asTrimmedString(payload.company);
  const email = asTrimmedString(payload.email);
  const message = asTrimmedString(payload.message);
  const locale = asTrimmedString(payload.locale);
  const pageUrl = asTrimmedString(payload.pageUrl);

  const lines = [
    'Animus landing — new lead',
    '',
    `Name: ${name || '—'}`,
    `Company: ${company || '—'}`,
    `Email: ${email || '—'}`,
    locale ? `Locale: ${locale}` : undefined,
    meta.ip ? `IP: ${meta.ip}` : undefined,
    meta.userAgent ? `UA: ${truncate(meta.userAgent, 220)}` : undefined,
    pageUrl ? `Page: ${truncate(pageUrl, 500)}` : undefined,
    `Time: ${meta.receivedAtIso}`,
    '',
    'Message:',
    message || '—',
  ].filter(Boolean) as string[];

  return truncate(lines.join('\n'), 3900);
}

async function sendTelegramMessage({
  botToken,
  chatId,
  threadId,
  text,
}: {
  botToken: string;
  chatId: string;
  threadId?: number;
  text: string;
}) {
  const endpoint = `https://api.telegram.org/bot${botToken}/sendMessage`;
  const payload: Record<string, unknown> = {
    chat_id: chatId,
    text,
    disable_web_page_preview: true,
  };
  if (typeof threadId === 'number' && Number.isFinite(threadId)) {
    payload.message_thread_id = threadId;
  }

  const response = await fetch(endpoint, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const text = await response.text().catch(() => '');
    throw new Error(`telegram sendMessage failed: ${response.status} ${text}`);
  }

  const data = (await response.json().catch(() => null)) as { ok?: boolean } | null;
  if (!data?.ok) {
    throw new Error('telegram sendMessage failed: non-ok response');
  }
}

export async function POST(request: Request) {
  const botToken = process.env.TELEGRAM_BOT_TOKEN?.trim();
  const chatId = process.env.TELEGRAM_CHAT_ID?.trim();
  const threadIdRaw = process.env.TELEGRAM_THREAD_ID?.trim();
  const threadId = threadIdRaw ? Number(threadIdRaw) : undefined;

  if (!botToken || !chatId) {
    return NextResponse.json(
      {
        ok: false,
        error: 'Telegram is not configured (missing TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID).',
      },
      { status: 503 },
    );
  }

  let payload: LeadPayload;
  try {
    payload = (await request.json()) as LeadPayload;
  } catch {
    return NextResponse.json({ ok: false, error: 'Invalid JSON body.' }, { status: 400 });
  }

  // Honeypot: bots tend to fill hidden fields.
  const website = asTrimmedString(payload.website);
  if (website) {
    return NextResponse.json({ ok: true }, { status: 200 });
  }

  const name = asTrimmedString(payload.name);
  const email = asTrimmedString(payload.email);
  const message = asTrimmedString(payload.message);

  if (!name || !email || !message) {
    return NextResponse.json(
      { ok: false, error: 'Missing required fields: name, email, message.' },
      { status: 400 },
    );
  }

  const receivedAtIso = new Date().toISOString();
  const forwardedFor = request.headers.get('x-forwarded-for');
  const ip = forwardedFor?.split(',')[0]?.trim() ?? request.headers.get('x-real-ip');
  const userAgent = request.headers.get('user-agent');

  const text = formatLeadMessage(
    { ...payload, name, email, message },
    { receivedAtIso, ip, userAgent },
  );

  try {
    await sendTelegramMessage({ botToken, chatId, threadId, text });
    return NextResponse.json({ ok: true }, { status: 200 });
  } catch (error) {
    console.error('[marketing] telegram lead send failed', error);
    return NextResponse.json({ ok: false, error: 'Failed to send message.' }, { status: 502 });
  }
}
