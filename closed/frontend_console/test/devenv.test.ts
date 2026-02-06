import { strict as assert } from 'node:assert';
import { test } from 'node:test';

import { buildDevEnvProxyUrl } from '@/lib/devenv';

test('buildDevEnvProxyUrl prefixes gateway base and experiments proxy', () => {
  process.env.NEXT_PUBLIC_GATEWAY_URL = 'https://gateway.example';
  const url = buildDevEnvProxyUrl('/devenv-sessions/abc/proxy');
  assert.equal(url, 'https://gateway.example/api/experiments/devenv-sessions/abc/proxy');
});

test('buildDevEnvProxyUrl avoids double prefix', () => {
  process.env.NEXT_PUBLIC_GATEWAY_URL = 'https://gateway.example';
  const url = buildDevEnvProxyUrl('/api/experiments/devenv-sessions/abc/proxy');
  assert.equal(url, 'https://gateway.example/api/experiments/devenv-sessions/abc/proxy');
});
