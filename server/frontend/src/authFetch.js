let refreshInFlight = null;

function clearSession() {
  for (const key of ['access', 'refresh', 'campaignsResponse', 'currentCampaignId']) localStorage.removeItem(key);
  window.location.replace('/');
}

async function refreshAccess(failedAccess) {
  const latestAccess = localStorage.getItem('access');
  if (latestAccess && latestAccess !== failedAccess) return latestAccess;

  const refresh = localStorage.getItem('refresh');
  if (!refresh) {
    clearSession();
    throw new Error('Session expired');
  }
  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      const response = await fetch('/rest/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${failedAccess}` },
        body: JSON.stringify({ token: refresh }),
        signal: AbortSignal.timeout(15000),
      });
      if (response.status === 400 || response.status === 401) {
        clearSession();
        throw new Error('Session expired');
      }
      if (!response.ok) throw new Error('Could not refresh session');
      const pair = await response.json();
      if (typeof pair?.access !== 'string' || !pair.access || typeof pair?.refresh !== 'string' || !pair.refresh) {
        clearSession();
        throw new Error('Invalid refresh response');
      }
      localStorage.setItem('access', pair.access);
      localStorage.setItem('refresh', pair.refresh);
      return pair.access;
    })().finally(() => { refreshInFlight = null; });
  }
  return refreshInFlight;
}

export async function authFetch(input, options = {}) {
  const access = localStorage.getItem('access');
  if (!access) {
    clearSession();
    throw new Error('Session required');
  }
  const send = token => {
    const headers = new Headers(options.headers);
    headers.set('Authorization', `Bearer ${token}`);
    return fetch(input, { ...options, headers });
  };
  let response = await send(access);
  if (response.status !== 401) return response;
  const newAccess = await refreshAccess(access);
  if (options.signal?.aborted) throw options.signal.reason;
  response = await send(newAccess);
  if (response.status === 401) {
    clearSession();
    throw new Error('Session expired');
  }
  return response;
}
