package cfg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cfg struct {
	Access     []byte
	Refresh    []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	Bind    string
	Port    int
	TlsKey  string
	TlsCert string
}

func Load() (*Cfg, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}
	path := filepath.Join(home, ".local", "share", "styx", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var file struct {
		Access     string `json:"jwt_access_secret"`
		Refresh    string `json:"jwt_refresh_secret"`
		AccessTTL  string `json:"access_ttl"`
		RefreshTTL string `json:"refresh_ttl"`
		Bind       string `json:"bind_address"`
		Port       int    `json:"port"`
		TlsKey     string `json:"tls_key_path"`
		TlsCert    string `json:"tls_cert_path"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	access, err := hex.DecodeString(file.Access)
	if err != nil || len(access) == 0 {
		return nil, fmt.Errorf("invalid jwt_access_secret")
	}
	refresh, err := hex.DecodeString(file.Refresh)
	if err != nil || len(refresh) == 0 {
		return nil, fmt.Errorf("invalid jwt_refresh_secret")
	}
	accessTTL, err := time.ParseDuration(file.AccessTTL)
	if err != nil || accessTTL <= 0 {
		return nil, fmt.Errorf("invalid access_ttl %q", file.AccessTTL)
	}
	refreshTTL, err := time.ParseDuration(file.RefreshTTL)
	if err != nil || refreshTTL <= 0 {
		return nil, fmt.Errorf("invalid refresh_ttl %q", file.RefreshTTL)
	}
	return &Cfg{Access: access, Refresh: refresh, AccessTTL: accessTTL, RefreshTTL: refreshTTL,
		Bind: file.Bind, Port: file.Port, TlsKey: file.TlsKey, TlsCert: file.TlsCert}, nil
}
