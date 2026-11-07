import assert from 'node:assert/strict';
import { afterEach, test } from 'node:test';
import { authFetch } from './authFetch.js';

const originalFetch = globalThis.fetch;

function setup() {
  const values = new Map([['access', 'old-access'], ['refresh', 'old-refresh'], ['currentCampaignId', 'campaign-a']]);
  globalThis.localStorage = {
    getItem: key => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, value),
    removeItem: key => values.delete(key),
  };
  let redirect = null;
  globalThis.window = { location: { replace: path => { redirect = path; } } };
  return { values, redirectedTo: () => redirect };
}

afterEach(() => {
  globalThis.fetch = originalFetch;
  delete globalThis.localStorage;
  delete globalThis.window;
});

test('401 refreshes once, stores the pair, and retries with the new access token', async () => {
  const state = setup();
  const calls = [];
  globalThis.fetch = async (url, options) => {
    calls.push({ url, authorization: new Headers(options.headers).get('Authorization'), body: options.body });
    if (url === '/rest/auth/refresh') return Response.json({ access: 'new-access', refresh: 'new-refresh' });
    return new Response('', { status: calls.length === 1 ? 401 : 200 });
  };

  const response = await authFetch('/rest/campaigns/list');
  assert.equal(response.status, 200);
  assert.deepEqual(calls.map(call => [call.url, call.authorization]), [
    ['/rest/campaigns/list', 'Bearer old-access'],
    ['/rest/auth/refresh', 'Bearer old-access'],
    ['/rest/campaigns/list', 'Bearer new-access'],
  ]);
  assert.deepEqual(JSON.parse(calls[1].body), { token: 'old-refresh' });
  assert.equal(state.values.get('access'), 'new-access');
  assert.equal(state.values.get('refresh'), 'new-refresh');
  assert.equal(state.redirectedTo(), null);
});

test('simultaneous 401 responses share one refresh', async () => {
  setup();
  let refreshes = 0;
  globalThis.fetch = async (url, options) => {
    if (url === '/rest/auth/refresh') {
      refreshes += 1;
      await new Promise(resolve => setTimeout(resolve, 5));
      return Response.json({ access: 'new-access', refresh: 'new-refresh' });
    }
    return new Response('', { status: new Headers(options.headers).get('Authorization') === 'Bearer old-access' ? 401 : 200 });
  };
  const responses = await Promise.all([authFetch('/one'), authFetch('/two')]);
  assert.deepEqual(responses.map(response => response.status), [200, 200]);
  assert.equal(refreshes, 1);
});

test('invalid refresh clears the session and returns to login', async () => {
  const state = setup();
  globalThis.fetch = async url => new Response('', { status: url === '/rest/auth/refresh' ? 401 : 401 });
  await assert.rejects(authFetch('/rest/campaigns/list'), /Session expired/);
  assert.equal(state.values.has('access'), false);
  assert.equal(state.values.has('refresh'), false);
  assert.equal(state.values.has('currentCampaignId'), false);
  assert.equal(state.redirectedTo(), '/login');
});

test('missing access token redirects to login without sending a request', async () => {
  const state = setup();
  state.values.delete('access');
  globalThis.fetch = () => { throw new Error('unexpected request'); };
  await assert.rejects(authFetch('/rest/campaigns/list'), /Session required/);
  assert.equal(state.redirectedTo(), '/login');
});
