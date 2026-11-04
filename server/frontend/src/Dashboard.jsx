import React from 'react';
import mark from './mark.svg';
import { CampaignProvider, useCampaign } from './CampaignContext.jsx';

function CampaignPicker() {
  const { campaigns, currentCampaignId, selectCampaign, status, retry } = useCampaign();
  const placeholder = status === 'loading' ? 'Loading campaigns…' : status === 'error' ? 'Campaigns unavailable' : campaigns.length ? 'Select a campaign' : 'No campaigns';
  return <div className="campaign-picker">
    <label htmlFor="campaign">Campaign</label>
    <select id="campaign" value={currentCampaignId} onChange={event => selectCampaign(event.target.value)} disabled={status !== 'ready' || !campaigns.length}>
      <option value="" disabled>{placeholder}</option>
      {campaigns.map(campaign => <option key={campaign.id} value={campaign.id}>{campaign.name}</option>)}
    </select>
    {status === 'error' && <div className="campaign-error" role="status">Could not load campaigns. <button onClick={retry}>Retry</button></div>}
  </div>;
}

export function Dashboard() {
  return <CampaignProvider>
    <div className="dashboard-shell">
      <header className="dashboard-topbar">
        <a className="brand" href="/dashboard" aria-label="Styx home"><img src={mark} alt="" width="30" height="30" /><span>styx</span></a>
        <nav className="top-nav" aria-label="Main navigation">
          <a className="nav-item" href="/dashboard" aria-current="page">Dashboard</a>
          {['Campaigns', 'Files', 'Builder', 'Settings'].map(name => <button className="nav-item" key={name} disabled title={`${name} is coming soon`}>{name}</button>)}
        </nav>
        <CampaignPicker />
      </header>
      <main className="dashboard-content" aria-label="Dashboard"><p>in development</p></main>
    </div>
  </CampaignProvider>;
}
