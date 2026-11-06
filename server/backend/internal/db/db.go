package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/argon2"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID TEXT NOT NULL,
		User TEXT NOT NULL,
		Pass TEXT NOT NULL
	);`,
	`CREATE TABLE IF NOT EXISTS campaigns(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT NOT NULL,
		Name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE TABLE IF NOT EXISTS signature(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID INTEGER NOT NULL,
		MetaHash TEXT NOT NULL,
		Salt TEXT,
		LicenseID TEXT NOT NULL,
		MachineID TEXT NOT NULL,
		ProductUUID TEXT NOT NULL,
		ProductSerial TEXT NOT NULL
	);`,
	`CREATE TABLE IF NOT EXISTS clients(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ClientName TEXT NOT NULL,
		ClientID   INTEGER NOT NULL,
		CampaignUUID TEXT NOT NULL,
		Hostname TEXT NOT NULL DEFAULT '',
		Domain TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(CampaignUUID) REFERENCES campaigns(uuid) ON DELETE CASCADE
	);`,
	`CREATE TABLE IF NOT EXISTS user_uploads(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		FileUUID  TEXT NOT NULL,
		FileName  TEXT NOT NULL,
		FileSize INTEGER NOT NULL,
		FileType TEXT NOT NULL DEFAULT '',
		HasReloc INTEGER NOT NULL,
		CampaignUUID TEXT NOT NULL,
		uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(CampaignUUID) REFERENCES campaigns(uuid) ON DELETE CASCADE
	);`,
}

type DB struct {
	conn *sql.DB
}

func New() (*DB, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".local", "share", "styx", "database", "styx.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil

}

func Setup(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	for _, val := range migrations {
		_, err := conn.Exec(val)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}
	rows, err := conn.Query(`PRAGMA table_info(signature)`)
	if err != nil {
		conn.Close()
		return nil, err
	}
	hasSalt := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, dataType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			conn.Close()
			return nil, err
		}
		if name == "Salt" {
			hasSalt = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !hasSalt {
		if _, err := conn.Exec(`ALTER TABLE signature ADD COLUMN Salt TEXT`); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil

}

func InsertUser(user, password, userID string, conn *sql.DB) error {
	const memory, iterations, parallelism, keyLength = 19 * 1024, 2, 1, 32
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
	q := `INSERT INTO users(User, Pass, UserID) VALUES(?,?,?)`

	_, err := conn.Exec(q, user, encoded, userID)
	return err
}

type LicenseInfo struct {
	User      string
	Pass      string
	LicenseID string
	MachineID string
	ProductID string
	SerialID  string
	MetaHash  string
	Salt      string
}

func InsertLicenseData(info LicenseInfo, db *sql.DB) error {
	q := `INSERT INTO signature(UserID, MetaHash, Salt, LicenseID, MachineID, ProductUUID, ProductSerial) VALUES(?,?,?,?,?,?,?)`

	_, err := db.Exec(q, "1211", info.MetaHash, info.Salt, info.LicenseID, info.MachineID, info.ProductID, info.SerialID)
	return err
}
