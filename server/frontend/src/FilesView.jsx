import React, { useEffect, useRef, useState } from 'react';
import { useCampaign } from './CampaignContext.jsx';
import { authFetch } from './authFetch.js';

function formatSize(bytes) {
  if (typeof bytes !== 'number' || bytes < 0) return '—';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function UploadDialog({ campaignId, onClose, onUploaded }) {
  const dialogRef = useRef(null);
  const [file, setFile] = useState(null);
  const [fileName, setFileName] = useState('');
  const [fileType, setFileType] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const dialog = dialogRef.current;
    dialog.showModal();
    return () => { if (dialog.open) dialog.close(); };
  }, []);

  function chooseFile(event) {
    const chosen = event.target.files?.[0] ?? null;
    setFile(chosen);
    setFileName(chosen?.name ?? '');
    setError('');
  }

  async function upload(event) {
    event.preventDefault();
    const name = fileName.trim();
    if (!file || !fileType || busy) return;
    if (!name || name === '.' || name === '..' || /[/\\\0]/.test(name)) {
      setError('Enter a file name without a path.');
      return;
    }
    setBusy(true);
    setError('');
    try {
      const start = await authFetch('/rest/uploads/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ file_name: name, campaign_uuid: campaignId, file_size: file.size, file_type: fileType }),
        signal: AbortSignal.timeout(15000),
      });
      if (!start.ok) throw new Error('Could not start the upload. Please try again.');
      const data = await start.json();
      if (!data?.id) throw new Error('The upload did not return a file ID.');
      const finish = await authFetch(`/rest/uploads/begin/${encodeURIComponent(data.id)}`, {
        method: 'PUT',
        body: file,
        signal: AbortSignal.timeout(120000),
      });
      if (!finish.ok) throw new Error('Could not finish the upload. Please try again.');
      onUploaded();
    } catch (cause) {
      setError(cause instanceof Error && cause.message.startsWith('Could not') ? cause.message : 'Upload failed. Please try again.');
    } finally {
      setBusy(false);
    }
  }

  return <dialog ref={dialogRef} className="upload-dialog" aria-labelledby="upload-title" onCancel={event => { if (busy) event.preventDefault(); else onClose(); }}>
    <h2 id="upload-title">Upload file</h2>
    <form onSubmit={upload}>
      <div className="upload-field"><label htmlFor="upload-file">Select file</label><input id="upload-file" type="file" onChange={chooseFile} required /></div>
      <div className="upload-field"><label htmlFor="upload-type">File type</label><select id="upload-type" value={fileType} onChange={event => setFileType(event.target.value)} required>
        <option value="" disabled>Select type</option><option value="exe">exe</option><option value="dll">dll</option><option value="shellcode">shellcode</option>
      </select></div>
      <div className="upload-field"><label htmlFor="upload-name">File name</label><input id="upload-name" value={fileName} onChange={event => setFileName(event.target.value)} placeholder="Choose a file first" required /></div>
      {error && <p className="upload-error" role="alert">{error}</p>}
      <div className="dialog-actions"><button type="button" disabled={busy} onClick={onClose}>Cancel</button><button className="upload-confirm" type="submit" disabled={busy || !file || !fileType || !fileName.trim()}>{busy ? 'Uploading…' : 'Upload'}</button></div>
    </form>
  </dialog>;
}

function DeleteFileDialog({ file, busy, error, onCancel, onConfirm }) {
  const dialogRef = useRef(null);
  useEffect(() => {
    const dialog = dialogRef.current;
    dialog.showModal();
    return () => { if (dialog.open) dialog.close(); };
  }, []);
  return <dialog ref={dialogRef} className="campaign-delete-dialog" aria-labelledby="delete-file-title" aria-describedby="delete-file-warning" onCancel={event => { if (busy) event.preventDefault(); else onCancel(); }}>
    <h2 id="delete-file-title">Delete {file.file_name}?</h2>
    <p id="delete-file-warning">This will remove the file from the campaign. Are you sure?</p>
    {error && <p className="delete-error" role="alert">{error}</p>}
    <div className="dialog-actions"><button type="button" autoFocus disabled={busy} onClick={onCancel}>No</button><button className="delete-confirm" type="button" disabled={busy} onClick={onConfirm}>{busy ? 'Deleting…' : 'Yes'}</button></div>
  </dialog>;
}

export function FilesView() {
  const { currentCampaignId, status: campaignStatus, retry: reloadCampaigns } = useCampaign();
  const [files, setFiles] = useState([]);
  const [status, setStatus] = useState('idle');
  const [requestedCampaignId, setRequestedCampaignId] = useState('');
  const [attempt, setAttempt] = useState(0);
  const [showUpload, setShowUpload] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  useEffect(() => { setShowUpload(false); setDeleteTarget(null); }, [currentCampaignId]);
  useEffect(() => {
    if (!currentCampaignId) return;
    const controller = new AbortController();
    setRequestedCampaignId(currentCampaignId);
    setFiles([]);
    setStatus('loading');
    async function load() {
      try {
        const response = await authFetch(`/rest/uploads/list/${encodeURIComponent(currentCampaignId)}`, {
          signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]),
        });
        if (!response.ok) throw new Error('File request failed');
        const data = await response.json();
        if (!data || (data.files !== null && !Array.isArray(data.files))) throw new Error('Invalid file list');
        if (!controller.signal.aborted) {
          setFiles(data.files ?? []);
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
  function uploaded() {
    setShowUpload(false);
    setAttempt(value => value + 1);
    reloadCampaigns();
  }

  async function deleteFile() {
    if (!deleteTarget || deleting) return;
    setDeleting(true);
    setDeleteError('');
    try {
      const response = await authFetch(`/rest/uploads/delete/${encodeURIComponent(deleteTarget.file_uuid)}`, {
        method: 'DELETE',
        signal: AbortSignal.timeout(15000),
      });
      if (!response.ok) throw new Error('Could not delete the file. Please try again.');
      setDeleteTarget(null);
      setAttempt(value => value + 1);
      reloadCampaigns();
    } catch {
      setDeleteError('Could not delete the file. Please try again.');
    } finally {
      setDeleting(false);
    }
  }

  return <main className="campaigns-view files-view" aria-labelledby="files-title">
    <header className="campaigns-heading files-heading"><h1 id="files-title">Files</h1><button type="button" disabled={!currentCampaignId || campaignStatus !== 'ready'} onClick={() => setShowUpload(true)}>Upload</button></header>
    {displayStatus === 'idle' && <p className="campaigns-message">choose campaign in top right</p>}
    {displayStatus === 'loading' && <p className="campaigns-message">Loading files…</p>}
    {displayStatus === 'error' && <p className="campaigns-message">Could not load files. <button className="text-action" type="button" onClick={() => setAttempt(value => value + 1)}>Retry</button></p>}
    {displayStatus === 'ready' && files.length === 0 && <p className="campaigns-message">No files uploaded for this campaign yet.</p>}
    {displayStatus === 'ready' && files.length > 0 && <table className="campaigns-table files-table">
      <thead><tr><th>File name</th><th>Type</th><th>Size</th><th>Uploaded at</th><th><span className="sr-only">Actions</span></th></tr></thead>
      <tbody>{files.map(file => <tr key={file.file_uuid}>
        <td className="campaign-name">{file.file_name || '—'}</td><td>{file.file_type || '—'}</td><td>{formatSize(file.file_size)}</td>
        <td>{file.uploaded_at && !Number.isNaN(Date.parse(file.uploaded_at)) ? new Date(file.uploaded_at).toLocaleString() : '—'}</td>
        <td className="file-actions"><button type="button" onClick={() => { setDeleteError(''); setDeleteTarget(file); }}>Delete</button></td>
      </tr>)}</tbody>
    </table>}
    {showUpload && currentCampaignId && <UploadDialog key={currentCampaignId} campaignId={currentCampaignId} onClose={() => setShowUpload(false)} onUploaded={uploaded} />}
    {deleteTarget && <DeleteFileDialog file={deleteTarget} busy={deleting} error={deleteError} onCancel={() => setDeleteTarget(null)} onConfirm={deleteFile} />}
  </main>;
}
