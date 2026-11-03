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
