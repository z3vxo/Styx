package setup

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteBaseConfigPreservesSecrets(t *testing.T) {
	dir := t.TempDir()
	if err := WriteBaseConfig(dir, os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"jwt_access_secret", "jwt_refresh_secret"} {
		value, ok := config[key].(string)
		decoded, err := hex.DecodeString(value)
		if !ok || err != nil || len(decoded) != 32 {
			t.Fatalf("%s is not a 32-byte secret", key)
		}
	}
	if config["jwt_access_secret"] == config["jwt_refresh_secret"] {
		t.Fatal("secrets are identical")
	}
	if config["access_ttl"] != "15m" || config["refresh_ttl"] != "168h" {
		t.Fatal("unexpected token lifetimes")
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != 0600 {
		t.Fatalf("permissions = %o", stat.Mode().Perm())
	}
	if err := WriteBaseConfig(dir, os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Fatal("existing config was overwritten")
	}
}
