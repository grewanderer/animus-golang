import { strict as assert } from 'node:assert';
import { afterEach, test } from 'node:test';

import { GatewayAPIError, gatewayFetchJSON } from '@/lib/gateway-client';

afterEach(() => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globalThis as any).fetch = undefined;
});

test('gatewayFetchJSON parses successful JSON response', async () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globalThis as any).fetch = async () =>
    new Response(JSON.stringify({ ok: true }), {
      status: 200,
      headers: { 'Content-Type': 'application/json', 'X-Request-Id': 'req-1' },
    });

  const result = await gatewayFetchJSON<{ ok: boolean }>('/healthz');
  assert.equal(result.ok, true);
});

test('gatewayFetchJSON surfaces ErrorResponse with request_id', async () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globalThis as any).fetch = async () =>
    new Response(JSON.stringify({ error: 'VALIDATION_FAILED', request_id: 'req-2' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });

  await assert.rejects(
    () => gatewayFetchJSON('/bad'),
    (err: unknown) =>
      err instanceof GatewayAPIError && err.code === 'VALIDATION_FAILED' && err.requestId === 'req-2',
  );
});
