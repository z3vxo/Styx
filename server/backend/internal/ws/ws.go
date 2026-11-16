package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	authmgr "styx/internal/auth"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

const writeTimeout = 5 * time.Second

// WS tracks dashboard WebSocket connections. Connections are not scoped to an
// agent; every authenticated dashboard tab receives registration events.
type WS struct {
	auth  *authmgr.Auth
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func New(auth *authmgr.Auth) *WS {
	return &WS{auth: auth, conns: make(map[*websocket.Conn]struct{})}
}

func (ws *WS) add(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.conns[conn] = struct{}{}
}

func (ws *WS) remove(conn *websocket.Conn) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	delete(ws.conns, conn)
}

// SendNewAgent broadcasts an agent registration to every dashboard tab.
func (ws *WS) SendNewAgent(agent any) error {
	body, err := json.Marshal(agent)
	if err != nil {
		return err
	}
	event, err := json.Marshal(WSEvent{Event: NewAgent, MessageID: uuid.NewString(), Body: body})
	if err != nil {
		return err
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()
	for conn := range ws.conns {
		writeCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		err := conn.Write(writeCtx, websocket.MessageText, event)
		cancel()
		if err != nil {
			delete(ws.conns, conn)
			_ = conn.Close(websocket.StatusInternalError, "send failed")
		}
	}
	return nil
}

func (ws *WS) HandleWS(w http.ResponseWriter, r *http.Request) {
	if _, err := ws.auth.Verify(accessToken(r)); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"styx"}})
	if err != nil {
		return
	}
	if conn.Subprotocol() != "styx" {
		_ = conn.Close(websocket.StatusPolicyViolation, "missing protocol")
		return
	}
	conn.SetReadLimit(64 << 10)

	ws.add(conn)
	defer func() {
		ws.remove(conn)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		if _, _, err := conn.Read(r.Context()); err != nil {
			return
		}
	}
}

func accessToken(r *http.Request) string {
	for _, protocol := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		if token, ok := strings.CutPrefix(strings.TrimSpace(protocol), "styx.jwt."); ok {
			return token
		}
	}
	return ""
}
