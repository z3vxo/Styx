package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	authmgr "styx/internal/auth"
	"styx/internal/db"

	"github.com/coder/websocket"
)

const writeTimeout = 5 * time.Second

type ConnInfo struct {
	TabID   string
	Conn    *websocket.Conn
	writeMu sync.Mutex
}

type WS struct {
	db    *db.DB
	auth  *authmgr.Auth
	Conns map[int]map[string]*ConnInfo
	mu    sync.RWMutex
}

func New(database *db.DB, auth *authmgr.Auth) *WS {
	return &WS{db: database, auth: auth, Conns: make(map[int]map[string]*ConnInfo)}
}

func (ws *WS) Add(clientID int, info *ConnInfo) {
	ws.mu.Lock()
	if ws.Conns[clientID] == nil {
		ws.Conns[clientID] = make(map[string]*ConnInfo)
	}
	previous := ws.Conns[clientID][info.TabID]
	ws.Conns[clientID][info.TabID] = info
	ws.mu.Unlock()
	if previous != nil && previous != info {
		_ = previous.Conn.Close(websocket.StatusNormalClosure, "replaced by new connection")
	}
}

func (ws *WS) Remove(clientID int, info *ConnInfo) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.Conns[clientID][info.TabID] != info {
		return
	}
	delete(ws.Conns[clientID], info.TabID)
	if len(ws.Conns[clientID]) == 0 {
		delete(ws.Conns, clientID)
	}
}

func (ws *WS) Send(clientID int, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ws.mu.RLock()
	connections := make([]*ConnInfo, 0, len(ws.Conns[clientID]))
	for _, info := range ws.Conns[clientID] {
		connections = append(connections, info)
	}
	ws.mu.RUnlock()
	if len(connections) == 0 {
		return fmt.Errorf("no websocket connections for client %d", clientID)
	}
	var sendErrors []error
	for _, info := range connections {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		info.writeMu.Lock()
		err := info.Conn.Write(ctx, websocket.MessageText, payload)
		info.writeMu.Unlock()
		cancel()
		if err != nil {
			sendErrors = append(sendErrors, fmt.Errorf("tab %s: %w", info.TabID, err))
			ws.Remove(clientID, info)
			_ = info.Conn.Close(websocket.StatusInternalError, "send failed")
		}
	}
	return errors.Join(sendErrors...)
}

func accessToken(r *http.Request) string {
	for _, protocol := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		if token, ok := strings.CutPrefix(strings.TrimSpace(protocol), "styx.jwt."); ok {
			return token
		}
	}
	return ""
}

func (ws *WS) HandleWS(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("client_id"), 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "invalid client ID", http.StatusBadRequest)
		return
	}
	tabID := r.URL.Query().Get("tab_id")
	if tabID == "" {
		http.Error(w, "missing tab ID", http.StatusBadRequest)
		return
	}
	if _, err := ws.auth.Verify(accessToken(r)); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	exists, err := ws.db.ClientExists(int32(id))
	if err != nil {
		http.Error(w, "could not verify client", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "client not found", http.StatusNotFound)
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
	info := &ConnInfo{TabID: tabID, Conn: conn}
	clientID := int(id)
	ws.Add(clientID, info)
	defer func() {
		ws.Remove(clientID, info)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		if _, _, err := conn.Read(r.Context()); err != nil {
			return
		}
	}
}
