package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authmgr "styx/internal/auth"

	"github.com/coder/websocket"
)

func testAuth() *authmgr.Auth {
	return &authmgr.Auth{
		AccessSec:  []byte("access-secret"),
		RefreshSec: []byte("refresh-secret"),
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}
}

func startServer(t *testing.T) (*WS, *authmgr.Auth, *httptest.Server) {
	t.Helper()
	auth := testAuth()
	handler := New(auth)
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/ws/connect", handler.HandleWS)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return handler, auth, server
}

func wsURL(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http") + "/rest/ws/connect"
}

func waitForConnections(t *testing.T, handler *WS, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var got int
	for time.Now().Before(deadline) {
		handler.mu.Lock()
		got = len(handler.conns)
		handler.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("websocket connections = %d, want %d", got, want)
}

func TestHandleWSBroadcastsNewAgent(t *testing.T) {
	handler, auth, server := startServer(t)
	tokens, err := auth.NewJwtPair("alice")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(server), &websocket.DialOptions{
		Subprotocols: []string{"styx", "styx.jwt." + tokens.Access},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	waitForConnections(t, handler, 1)

	payload := struct {
		CampaignID string `json:"campaign_id"`
		Name       string `json:"name"`
		ClientID   int32  `json:"client_id"`
	}{CampaignID: "campaign-a", Name: "Nyx-Tartarus", ClientID: 42}
	if err := handler.SendNewAgent(payload); err != nil {
		t.Fatal(err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var event WSEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatal(err)
	}
	if event.Event != NewAgent || event.MessageID == "" {
		t.Fatalf("event = %+v", event)
	}
	var body struct {
		CampaignID string `json:"campaign_id"`
		Name       string `json:"name"`
		ClientID   int32  `json:"client_id"`
	}
	if err := json.Unmarshal(event.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.CampaignID != payload.CampaignID || body.Name != payload.Name || body.ClientID != payload.ClientID {
		t.Fatalf("body = %+v", body)
	}
}

func TestHandleWSRequiresToken(t *testing.T) {
	_, _, server := startServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, wsURL(server), &websocket.DialOptions{
		Subprotocols: []string{"styx"},
	})
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
		t.Fatal("connection without a token was accepted")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("response = %+v, err = %v", resp, err)
	}
}
