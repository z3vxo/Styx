package setup

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"styx/internal/db"
	"styx/internal/modules"
)

func Run() (*sql.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	uid, gid := os.Getuid(), os.Getgid()
	if name := os.Getenv("SUDO_USER"); os.Geteuid() == 0 && name != "" {
		u, err := user.Lookup(name)
		if err != nil {
			return nil, fmt.Errorf("look up invoking user %q: %w", name, err)
		}
		home = u.HomeDir
		uid, err = strconv.Atoi(u.Uid)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		gid, err = strconv.Atoi(u.Gid)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID: %w", err)
		}
	}

	baseDir := filepath.Join(home, ".local", "share", "styx")

	for _, dir := range []string{filepath.Join(home, ".local"), filepath.Join(home, ".local", "share"), baseDir} {
		if err := os.Mkdir(dir, 0755); err != nil {
			if !os.IsExist(err) {
				return nil, err
			}
		} else if os.Geteuid() == 0 {
			if err := os.Chown(dir, uid, gid); err != nil {
				return nil, err
			}
		}
	}

	dirs := []string{
		"database",
		"downloads",
		"user-files",
		"modules",
		"modules/file-system",
		"modules/ad-enum",
		"modules/pe-loader",
	}

	for _, entry := range dirs {
		dir := filepath.Join(baseDir, entry)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		if os.Geteuid() == 0 {
			if err := os.Chown(dir, uid, gid); err != nil {
				return nil, err
			}
		}
	}

	modulesPath := filepath.Join(baseDir, "modules", "modules.json")
	if _, err := os.Stat(modulesPath); errors.Is(err, os.ErrNotExist) {
		data, err := json.MarshalIndent(modules.Default().Modules, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(modulesPath, data, 0644); err != nil {
			return nil, err
		}
		if os.Geteuid() == 0 {
			if err := os.Chown(modulesPath, uid, gid); err != nil {
				return nil, err
			}
		}
	}

	dbPath := filepath.Join(baseDir, "database", "styx.db")

	conn, err := db.Setup(dbPath)
	if err != nil {
		return nil, err
	}
	if os.Geteuid() == 0 {
		if err := os.Chown(baseDir, uid, gid); err != nil {
			conn.Close()
			return nil, err
		}
		if err := os.Chown(dbPath, uid, gid); err != nil {
			conn.Close()
			return nil, err
		}
	}
	if err := WriteBaseConfig(baseDir, uid, gid); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write base config: %w", err)
	}

	fmt.Printf("[*] Created file and setup db %s\n", baseDir)

	return conn, nil
}

func Register(username, password, domain string, conn *sql.DB) error {
	user := username
	pass := password
	backend := domain
	if user == "" {
		user = os.Getenv("STYX_USER")
		if user == "" {
			return fmt.Errorf("Error: username must be provided by either env vars or cmd flag")
		}
	}
	if pass == "" {
		pass = os.Getenv("STYX_PASS")
		if pass == "" {
			return fmt.Errorf("Error: password must be provided by either env vars or cmd flag")
		}
	}

	if backend == "" {
		backend = os.Getenv("STYX_URL")
		if backend == "" {
			return fmt.Errorf("Error: backend must be provided by either env vars or cmd flag")
		}
	}

	data, err := CollectRegitrationInfo(user, pass)
	if err != nil {
		return err
	}

	userID, err := RegisterToServer(backend, data)
	if err != nil {
		return fmt.Errorf("register with server: %w", err)
	}
	fmt.Printf("[*] Registered user %s\n", userID)

	if err := db.InsertUser(user, pass, userID, conn); err != nil {
		return err
	}

	if err := db.InsertLicenseData(data, conn); err != nil {
		return err
	}
	conn.Close()

	return nil
}

func WriteBaseConfig(baseDir string, uid, gid int) error {
	path := filepath.Join(baseDir, "config.json")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	complete := false
	defer func() {
		f.Close()
		if !complete {
			os.Remove(path)
		}
	}()
	if os.Geteuid() == 0 {
		if err := f.Chown(uid, gid); err != nil {
			return err
		}
	}
	secret := func() (string, error) {
		data := make([]byte, 32)
		if _, err := rand.Read(data); err != nil {
			return "", err
		}
		return hex.EncodeToString(data), nil
	}
	accessSecret, err := secret()
	if err != nil {
		return err
	}
	refreshSecret, err := secret()
	if err != nil {
		return err
	}
	config := struct {
		BindAddress      string `json:"bind_address"`
		Port             int    `json:"port"`
		JWTAccessSecret  string `json:"jwt_access_secret"`
		JWTRefreshSecret string `json:"jwt_refresh_secret"`
		AccessTTL        string `json:"access_ttl"`
		RefreshTTL       string `json:"refresh_ttl"`
		TLSCertPath      string `json:"tls_cert_path"`
		TLSKeyPath       string `json:"tls_key_path"`
	}{"127.0.0.1", 8080, accessSecret, refreshSecret, "15m", "168h", "", ""}
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	complete = true
	return nil
}
