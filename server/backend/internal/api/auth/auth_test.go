package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	authmgr "styx/internal/auth"
	"styx/internal/cfg"
	"testing"
	"time"
)

func TestRefreshReturnsNewTokenPair(t *testing.T) {
	service := authmgr.New(&cfg.Cfg{Access: []byte("access-secret"), Refresh: []byte("refresh-secret"), AccessTTL: time.Minute, RefreshTTL: time.Hour})
	old, err := service.NewJwtPair("tester")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"token":"`+old.Refresh+`"}`))
	request.Header.Set("Authorization", "Bearer "+old.Access)
	response := httptest.NewRecorder()
	New(nil, service).Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", response.Code, response.Body.String())
	}
	var pair authmgr.Tokens
	if err := json.Unmarshal(response.Body.Bytes(), &pair); err != nil {
		t.Fatal(err)
	}
	if pair.Access == "" || pair.Refresh == "" {
		t.Fatalf("incomplete token pair: %+v", pair)
	}
	if _, err := service.Verify(pair.Access); err != nil {
		t.Fatalf("new access token is invalid: %v", err)
	}
	if _, err := service.VerifyRefresh(pair.Refresh); err != nil {
		t.Fatalf("new refresh token is invalid: %v", err)
	}
}

func TestRefreshRejectsInvalidInput(t *testing.T) {
	service := authmgr.New(&cfg.Cfg{Access: []byte("access-secret"), Refresh: []byte("refresh-secret"), AccessTTL: time.Minute, RefreshTTL: time.Hour})
	for _, tc := range []struct {
		body string
		want int
	}{
		{`not json`, http.StatusBadRequest},
		{`{}`, http.StatusBadRequest},
		{`{"token":"invalid"}`, http.StatusUnauthorized},
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(tc.body))
		New(nil, service).Routes().ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Errorf("body %q: status = %d, want %d", tc.body, response.Code, tc.want)
		}
	}
}
