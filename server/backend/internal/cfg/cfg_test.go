package cfg_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"styx/internal/cfg"
	"styx/internal/setup"
)

func TestLoadSetupConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".local", "share", "styx")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := setup.WriteBaseConfig(dir, os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
	config, err := cfg.Load()
	if err != nil {
		t.Fatal(err)
	}
	if config.Bind != "127.0.0.1" || config.Port != 8080 || len(config.Access) != 32 || len(config.Refresh) != 32 || config.AccessTTL != 15*time.Minute || config.RefreshTTL != 7*24*time.Hour {
		t.Fatal("loaded config does not match setup defaults")
	}
}
