import React, { useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import mark from './mark.svg';
import './styles.css';
import { Dashboard } from './Dashboard.jsx';

function Login() {
  const [visible, setVisible] = useState(false);
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState({ text: '', success: false });
  const busy = useRef(false);
  async function signIn(event) {
    event.preventDefault();
    if (busy.current) return;
    const form = event.currentTarget;
    busy.current = true;
    setPending(true);
    setMessage({ text: '', success: false });
    try {
      const response = await fetch('/rest/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: form.elements.username.value, password: form.elements.password.value }),
        signal: AbortSignal.timeout(15000),
      });
      if (!response.ok) {
        if (response.status === 401 || response.status === 500) throw new Error('Unable to sign in. Check your username and password, then try again.');
        throw new Error('Styx is unavailable right now. Please try again shortly.');
      }
      const tokens = await response.json();
      if (!tokens.access || !tokens.refresh) throw new Error('Sign-in could not be completed. Please try again.');
      localStorage.setItem('access', tokens.access);
      localStorage.setItem('refresh', tokens.refresh);
      localStorage.removeItem('campaignsResponse');
      localStorage.removeItem('currentCampaignId');
      form.elements.password.value = '';
      window.location.assign('/dashboard');
    } catch (error) {
      setMessage({ text: error.name === 'TimeoutError' || error instanceof TypeError
        ? 'Could not reach Styx. Please try again shortly.' : error.message, success: false });
    } finally {
      busy.current = false;
      setPending(false);
    }
  }
  return (<>
<header className="site-header">
    <a className="brand" href="/" aria-label="Styx home"><img src={mark} alt="" width="30" height="30" /><span>styx</span></a>
  </header>
  <main>
    <section className="login" aria-labelledby="login-title">
      <div className="login-heading">
        <div className="entry-mark" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none"><path d="M14 4h5v16h-5M3 12h12m-4-4 4 4-4 4"/></svg></div>
        <h1 id="login-title">Welcome back.</h1>
        <p>Sign in to your Styx workspace.</p>
      </div>
      <form id="login-form" onSubmit={signIn} aria-busy={pending}>
        <div className="field"><label htmlFor="username">Username</label><input id="username" name="username" autoComplete="username" autoCapitalize="none" spellCheck={false} placeholder="Enter your username" required /></div>
        <div className="field"><label htmlFor="password">Password</label><div className="password-field"><input id="password" name="password" type={visible ? "text" : "password"} autoComplete="current-password" placeholder="Enter your password" required /><button className="reveal" id="reveal-password" type="button" aria-label={visible ? "Hide password" : "Show password"} aria-pressed={visible} onClick={() => setVisible(!visible)}><svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6Z"/><circle cx="12" cy="12" r="2.5"/><path className="eye-slash" d="m4 4 16 16"/></svg></button></div></div>
        <p id="form-message" className={`form-message${message.success ? " success" : ""}`} role="status" aria-live="polite" hidden={!message.text}>{message.text}</p>
        <button className="submit" type="submit" disabled={pending}><span>{pending ? "Signing in…" : "Sign in"}</span><svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M5 12h14m-5-5 5 5-5 5"/></svg></button>
      </form>
    </section>
  </main>
  <footer><span>Styx workspace</span></footer>
  </>);
}

const dashboardRoute = window.location.pathname === '/dashboard' || window.location.pathname === '/dashboard/';
createRoot(document.getElementById('root')).render(dashboardRoute ? <Dashboard /> : <Login />);
