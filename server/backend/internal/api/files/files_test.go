package files

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	authmgr "styx/internal/auth"
	"styx/internal/cfg"
	"styx/internal/db"
	"testing"
	"time"
)

func TestDeleteUploadRemovesDatabaseRowAndStoredFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	databaseDir := filepath.Join(home, ".local", "share", "styx", "database")
	if err := os.MkdirAll(databaseDir, 0755); err != nil {
		t.Fatal(err)
	}
	conn, err := db.Setup(filepath.Join(databaseDir, "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	const fileID = "0194e888-26fa-7a2a-b33a-e1dca858a531"
	if _, err := conn.Exec(`INSERT INTO user_uploads (FileUUID, FileName, FileSize, HasReloc, CampaignUUID) VALUES (?, ?, ?, ?, ?)`, fileID, "agent.exe", 4, 1, "campaign-a"); err != nil {
		t.Fatal(err)
	}
	fileDir := filepath.Join(home, ".local", "share", "styx", "user-files", fileID)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fileDir, "agent.exe"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	database, err := db.New()
	if err != nil {
		t.Fatal(err)
	}
	authService := authmgr.New(&cfg.Cfg{Access: []byte("access-secret"), Refresh: []byte("refresh-secret"), AccessTTL: time.Minute, RefreshTTL: time.Hour})
	tokens, err := authService.NewJwtPair("tester")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/delete/"+fileID, nil)
	request.Header.Set("Authorization", "Bearer "+tokens.Access)
	response := httptest.NewRecorder()
	New(database, authService).Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", response.Code, response.Body.String())
	}
	if _, err := os.Stat(fileDir); !os.IsNotExist(err) {
		t.Fatalf("stored file directory still exists: %v", err)
	}
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM user_uploads WHERE FileUUID = ?`, fileID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("upload row still exists: %d", count)
	}
}
