import React, { createContext, useContext, useEffect, useState } from 'react';
import { authFetch } from './authFetch.js';

const AgentSocketContext = createContext(null);
const retryDelayMs = 5000;

export function AgentSocketProvider({ children }) {
  const [status, setStatus] = useState('connecting');
  const [lastAgent, setLastAgent] = useState(null);
  const [outputsByAgent, setOutputsByAgent] = useState({});

  useEffect(() => {
    let disposed = false;
    let socket = null;
    let retryTimer = null;

    function scheduleReconnect() {
      if (disposed) return;
      retryTimer = window.setTimeout(connect, retryDelayMs);
    }

    async function connect() {
      if (disposed) return;
      setStatus('connecting');
      try {
        // Run a protected request first so the normal refresh flow can renew an
        // expired access token before it is offered as the socket subprotocol.
        const response = await authFetch('/rest/campaigns/list', {
          signal: AbortSignal.timeout(15000),
        });
        if (!response.ok) throw new Error('Could not authorize the connection');
        if (disposed) return;

        const access = localStorage.getItem('access');
        if (!access) throw new Error('Session required');
        const url = new URL('/rest/ws/connect', window.location.href);
        url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';

        const next = new WebSocket(url, ['styx', `styx.jwt.${access}`]);
        socket = next;
        next.onopen = () => {
          if (!disposed) setStatus('connected');
        };
        next.onmessage = event => {
          let message;
          try { message = JSON.parse(event.data); } catch { return; }
          const body = message.body;
          if (!body) return;
          const messageId = message.message_id ?? '';
          if (message.event === 'agent.New') {
            setLastAgent({ ...body, messageId });
            return;
          }
          if (message.event === 'task.Output' && body.agent_id != null) {
            setOutputsByAgent(previous => {
              const outputs = [...(previous[body.agent_id] ?? []), { ...body, messageId }];
              if (outputs.length > 200) outputs.splice(0, outputs.length - 200);
              return { ...previous, [body.agent_id]: outputs };
            });
          }
        };
        next.onerror = () => {
          if (!disposed) setStatus('error');
        };
        next.onclose = () => {
          if (socket === next) socket = null;
          if (disposed) return;
          setStatus('reconnecting');
          scheduleReconnect();
        };
      } catch {
        if (!disposed) {
          setStatus('error');
          scheduleReconnect();
        }
      }
    }

    connect();

    return () => {
      disposed = true;
      if (retryTimer) window.clearTimeout(retryTimer);
      if (!socket) return;
      const closed = socket;
      socket = null;
      closed.onopen = closed.onmessage = closed.onerror = closed.onclose = null;
      closed.close();
    };
  }, []);

  return <AgentSocketContext.Provider value={{ status, lastAgent, outputsByAgent }}>{children}</AgentSocketContext.Provider>;
}

export function useAgentSocket() {
  const context = useContext(AgentSocketContext);
  if (!context) throw new Error('useAgentSocket must be used within AgentSocketProvider');
  return context;
}
