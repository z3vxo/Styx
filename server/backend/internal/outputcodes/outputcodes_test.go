package outputcodes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMessages(t *testing.T) {
	const data = `{
		"version": 1,
		"core": {"177": "MODULE: Agent successfully loaded module"},
		"win32": {"5": "Access is denied."},
		"modules": {
			"file-system": {
				"range": "200-299",
				"codes": {"200": "CD: Agent successfully changed directories"}
			}
		}
	}`

	table, err := Parse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if message, ok := table.Message(NamespaceCore, 0xb1); !ok || message != "MODULE: Agent successfully loaded module" {
		t.Fatalf("core message = %q, ok = %v", message, ok)
	}
	if message, ok := table.Message(NamespaceModule, 200); !ok || message != "CD: Agent successfully changed directories" {
		t.Fatalf("module message = %q, ok = %v", message, ok)
	}
	if message, ok := table.Message(NamespaceWin32, 5); !ok || message != "Access is denied." {
		t.Fatalf("win32 message = %q, ok = %v", message, ok)
	}
	if _, ok := table.Message(NamespaceModule, 999); ok {
		t.Fatal("unknown module message resolved")
	}
	if _, ok := table.Message(NamespaceCore, 999); ok {
		t.Fatal("unknown core message resolved")
	}
}

func TestParseAcceptsHexAndRanges(t *testing.T) {
	const data = `{
		"version": 1,
		"core": {"0xB1": "core"},
		"win32": {"0x05": "denied"},
		"modules": {
			"file-system": {
				"range": "0xC8-0x12B",
				"codes": {"0xD1": "module"}
			}
		}
	}`

	table, err := parseTable([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if message, ok := table.Message(NamespaceCore, 177); !ok || message != "core" {
		t.Fatalf("core message = %q, ok = %v", message, ok)
	}
	if message, ok := table.Message(NamespaceWin32, 5); !ok || message != "denied" {
		t.Fatalf("win32 message = %q, ok = %v", message, ok)
	}
	if message, ok := table.Message(NamespaceModule, 209); !ok || message != "module" {
		t.Fatalf("module message = %q, ok = %v", message, ok)
	}
}

func TestParseRejectsInvalidCodes(t *testing.T) {
	for name, data := range map[string]string{
		"outside range": `{
			"version": 1,
			"modules": {"file-system": {"range": "200-299", "codes": {"300": "bad"}}}
		}`,
		"overlapping ranges": `{
			"version": 1,
			"modules": {
				"first": {"range": "200-299", "codes": {"200": "first"}},
				"second": {"range": "299-399", "codes": {"299": "second"}}
			}
		}`,
		"bad range": `{
			"version": 1,
			"modules": {"file-system": {"range": "299-200", "codes": {"200": "bad"}}}
		}`,
		"bad code": `{
			"version": 1,
			"core": {"nope": "bad"}
		}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseTable([]byte(data)); err == nil {
				t.Fatal("invalid registry accepted")
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := Load(); err == nil {
		t.Fatal("missing codes.json accepted")
	}
}

func TestLoadReadsInstalledFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".local", "share", "styx", "modules", "codes.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	const data = `{"version":1,"core":{"42":"installed"}}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	table, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if message, ok := table.Message(NamespaceCore, 42); !ok || message != "installed" {
		t.Fatalf("message = %q, ok = %v", message, ok)
	}
}
