import React, { createContext, useContext, useEffect, useState } from 'react';
import { authFetch } from './authFetch.js';

const CampaignContext = createContext(null);

export function CampaignProvider({ children }) {
  const [campaignsResponse, setCampaignsResponse] = useState(null);
  const [currentCampaignId, setCurrentCampaignId] = useState('');
  const [status, setStatus] = useState('loading');
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    async function load() {
      setStatus('loading');
      try {
        const response = await authFetch('/rest/campaigns/list', {
          signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]),
        });
        if (!response.ok) throw new Error('Campaign request failed');
        const data = await response.json();
        if (!data || (data.campaigns !== null && !Array.isArray(data.campaigns))) throw new Error('Invalid campaign list');
        const campaigns = data.campaigns ?? [];
        if (campaigns.some(c => !c || typeof c.id !== 'string' || typeof c.name !== 'string')) throw new Error('Invalid campaign');
        if (!active) return;
        const saved = localStorage.getItem('currentCampaignId');
        const selected = campaigns.some(c => c.id === saved) ? saved : '';
        localStorage.setItem('campaignsResponse', JSON.stringify(data));
        if (!selected) localStorage.removeItem('currentCampaignId');
        setCampaignsResponse(data);
        setCurrentCampaignId(selected);
        setStatus('ready');
      } catch {
        if (active) setStatus('error');
      }
    }
    load();
    return () => { active = false; controller.abort(); };
  }, [attempt]);

  const campaigns = campaignsResponse?.campaigns ?? [];
  function selectCampaign(id) {
    if (!campaigns.some(c => c.id === id)) return;
    try {
      localStorage.setItem('currentCampaignId', id);
      setCurrentCampaignId(id);
    } catch { setStatus('error'); }
  }

  return <CampaignContext.Provider value={{ campaignsResponse, campaigns, currentCampaignId, selectCampaign, status, retry: () => setAttempt(n => n + 1) }}>
    {children}
  </CampaignContext.Provider>;
}

export function useCampaign() {
  const context = useContext(CampaignContext);
  if (!context) throw new Error('useCampaign must be used within CampaignProvider');
  return context;
}
