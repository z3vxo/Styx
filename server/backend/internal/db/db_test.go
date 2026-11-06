package db

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestInsertUserHashesPassword(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := InsertUser("alice", "password", "id", conn); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := conn.QueryRow(`SELECT Pass FROM users WHERE User = ?`, "alice").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(stored, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		t.Fatal("invalid Argon2id encoding")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		t.Fatal(err)
	}
	want := argon2.IDKey([]byte("password"), salt, 2, 19*1024, 1, 32)
	if parts[5] != base64.RawStdEncoding.EncodeToString(want) {
		t.Fatal("password hash does not match")
	}
}

func TestVerifyAuth(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := InsertUser("alice", "correct", "id", conn); err != nil {
		t.Fatal(err)
	}
	database := &DB{conn: conn}
	if err := database.VerifyAuth("alice", "correct"); err != nil {
		t.Fatal(err)
	}
	if err := database.VerifyAuth("alice", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	if err := database.VerifyAuth("unknown", "correct"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown user: %v", err)
	}
	if _, err := conn.Exec(`UPDATE users SET Pass = ? WHERE User = ?`, "$argon2id$v=19$m=0,t=1,p=0$bad$bad", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := database.VerifyAuth("alice", "correct"); err == nil {
		t.Fatal("accepted malformed hash")
	}
}

func TestSetupMigratesExistingSignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "styx.db")
	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = old.Exec(`CREATE TABLE signature (id INTEGER PRIMARY KEY AUTOINCREMENT, UserID INTEGER NOT NULL, MetaHash TEXT NOT NULL, LicenseID TEXT NOT NULL, MachineID TEXT NOT NULL, ProductUUID TEXT NOT NULL, ProductSerial TEXT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	conn, err := Setup(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	info := LicenseInfo{Salt: "example-salt", MetaHash: "hash", LicenseID: "license", MachineID: "machine", ProductID: "product", SerialID: "serial"}
	if err := InsertLicenseData(info, conn); err != nil {
		t.Fatal(err)
	}
	var salt string
	if err := conn.QueryRow(`SELECT Salt FROM signature`).Scan(&salt); err != nil {
		t.Fatal(err)
	}
	if salt != info.Salt {
		t.Fatalf("salt = %q, want %q", salt, info.Salt)
	}
}

func TestListCampaignsCountsClientsAndFiles(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for _, id := range []string{"campaign-a", "campaign-b", "campaign-empty"} {
		if _, err := conn.Exec(`INSERT INTO campaigns (Name, uuid) VALUES (?, ?)`, id, id); err != nil {
			t.Fatal(err)
		}
	}
	for i, campaignID := range []string{"campaign-a", "campaign-a", "campaign-b"} {
		if _, err := conn.Exec(`INSERT INTO clients (ClientName, ClientID, CampaignUUID) VALUES (?, ?, ?)`, "client", i+1, campaignID); err != nil {
			t.Fatal(err)
		}
	}
	for _, upload := range []struct{ id, campaignID string }{
		{"file-a", "campaign-a"},
		{"file-b", "campaign-b"},
		{"file-c", "campaign-b"},
	} {
		if _, err := conn.Exec(`INSERT INTO user_uploads (FileUUID, FileName, FileSize, HasReloc, CampaignUUID) VALUES (?, ?, ?, ?, ?)`,
			upload.id, "test.txt", 4, 1, upload.campaignID); err != nil {
			t.Fatal(err)
		}
	}
	result, err := (&DB{conn: conn}).ListCampaigns()
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 3 {
		t.Fatalf("campaign total = %d, want 3", result.Total)
	}
	want := map[string]struct{ clients, files int }{
		"campaign-a":     {2, 1},
		"campaign-b":     {1, 2},
		"campaign-empty": {0, 0},
	}
	for _, campaign := range result.Entry {
		counts, ok := want[campaign.Id]
		if !ok {
			t.Errorf("unexpected campaign %s", campaign.Id)
			continue
		}
		if campaign.TotalClients != counts.clients || campaign.TotalFiles != counts.files {
			t.Errorf("%s totals = clients %d, files %d; want clients %d, files %d",
				campaign.Id, campaign.TotalClients, campaign.TotalFiles, counts.clients, counts.files)
		}
		delete(want, campaign.Id)
	}
	if len(want) != 0 {
		t.Fatalf("missing campaigns: %v", want)
	}
}

func TestDeleteCampaignRemovesRelatedRecords(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for _, id := range []string{"campaign-a", "campaign-b"} {
		if _, err := conn.Exec(`INSERT INTO campaigns (Name, uuid) VALUES (?, ?)`, id, id); err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(`INSERT INTO clients (ClientName, ClientID, CampaignUUID) VALUES (?, ?, ?)`, "client", 1, id); err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Exec(`INSERT INTO user_uploads (FileUUID, FileName, FileSize, HasReloc, CampaignUUID) VALUES (?, ?, ?, ?, ?)`, id+"-file", "file.txt", 4, 1, id); err != nil {
			t.Fatal(err)
		}
	}
	fileIDs, err := (&DB{conn: conn}).DeleteCampaign("campaign-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(fileIDs) != 1 || fileIDs[0] != "campaign-a-file" {
		t.Fatalf("deleted file IDs = %v", fileIDs)
	}
	for _, table := range []string{"campaigns", "clients", "user_uploads"} {
		for _, tc := range []struct {
			id   string
			want int
		}{{"campaign-a", 0}, {"campaign-b", 1}} {
			column := "CampaignUUID"
			if table == "campaigns" {
				column = "uuid"
			}
			var count int
			if err := conn.QueryRow(`SELECT COUNT(*) FROM `+table+` WHERE `+column+` = ?`, tc.id).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != tc.want {
				t.Errorf("%s rows for %s = %d, want %d", table, tc.id, count, tc.want)
			}
		}
	}
}
