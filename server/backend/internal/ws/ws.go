package ws

import (
	"context"
	"encoding/json"
	"net/http"
	authmgr "styx/internal/auth"
	"styx/internal/db"
	"sync"

	"github.com/coder/websocket"
	"github.com/go-chi/chi"
)

type WS struct {
	db    *db.DB
	auth  *authmgr.Auth
	Conns map[string]*websocket.Conn
	mu    sync.RWMutex
}

type WSEvent struct {
	Event   string          `json:"event"`
	Message string          `json:"message"`
	Body    json.RawMessage `json:"body,omitempty"`
}

type NewTaskEventReq struct {
	Cmd        string `json:"cmd"`
	Args       []Arg  `json:"args"`
	ClientID   string `json:"client_id"`
	CampaignID string `json:"campaign_id"`
}

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func New(database *db.DB, a *authmgr.Auth) *WS {
	return &WS{db: database, auth: a, Conns: make(map[string]*websocket.Conn)}
}

func (ws *WS) Add(id string, conn *websocket.Conn) {
	ws.mu.Lock()
	ws.Conns[id] = conn
	ws.mu.Unlock()
}

func (ws *WS) Remove(id string) {
	ws.mu.Lock()
	delete(ws.Conns, id)
	ws.mu.Unlock()
}

func (ws *WS) Reply(conn *websocket.Conn, data any) {
	resp, err := json.Marshal(data)
	if err != nil {
		return
	}
	conn.Write(context.Background(), websocket.MessageText, resp)
}

func (ws *WS) Handle(conn *websocket.Conn, msg []byte) {
	var req WSEvent
	if err := json.Unmarshal(msg, &req); err != nil {
		ws.Reply(conn, WSEvent{Event: "request.invalid_json", Message: "failed decoding json"})
		return
	}

}

func (ws *WS) HandleWS(w http.ResponseWriter, r *http.Request) {
	ClientID := chi.URLParam(r, "id")
	if ClientID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provided clientID in url"})
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed upgrading to WS"})
		return
	}

	defer func() {
		ws.Remove(ClientID)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}()

	ws.Add(ClientID, conn)

	for {
		_, msg, err := conn.Read(r.Context())
		if err != nil {
			break
		}
		ws.Handle(conn, msg)
	}

}
