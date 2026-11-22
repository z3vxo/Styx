package modules

import (
	"encoding/json"
	"testing"
)

func TestDefaultResolvesModulesAndCommands(t *testing.T) {
	registry := Default()
	module, ok := registry.Module("file-system")
	if !ok || module.ID != 1 {
		t.Fatalf("file-system = %+v, ok = %v", module, ok)
	}
	name, resolved, code, ok := registry.ResolveCommand("ls")
	if !ok || name != "file-system" || resolved.ID != 1 || code != 1 {
		t.Fatalf("ls = %q %+v %d %v", name, resolved, code, ok)
	}
	if _, _, _, ok := registry.ResolveCommand("missing"); ok {
		t.Fatal("unknown command resolved")
	}
}

func TestCommandJSONAcceptsLegacyAndDescribedForms(t *testing.T) {
	const registryJSON = `{
		"file-system": {
			"abi": 1,
			"id": 1,
			"file": "file-system.mod",
			"commands": {
				"ls": 1,
				"pwd": {"code": 3},
				"cat": {
					"code": 4,
					"usage": "cat <path>",
					"args": [{"name": "path", "type": "string", "required": true}]
				}
			}
		}
	}`

	var parsed map[string]Modules
	if err := json.Unmarshal([]byte(registryJSON), &parsed); err != nil {
		t.Fatal(err)
	}

	module := parsed["file-system"]
	if module.ABI != 1 || module.Commands["ls"].Code != 1 {
		t.Fatalf("legacy command = %+v", module.Commands["ls"])
	}

	cat := module.Commands["cat"]
	if cat.Code != 4 || cat.Usage != "cat <path>" || len(cat.Args) != 1 {
		t.Fatalf("cat = %+v", cat)
	}
	if arg := cat.Args[0]; arg.Name != "path" || arg.Type != "string" || !arg.Required {
		t.Fatalf("cat arg = %+v", arg)
	}
	if pwd := module.Commands["pwd"]; pwd.Args == nil || len(pwd.Args) != 0 {
		t.Fatalf("pwd args = %#v", pwd.Args)
	}
}

func TestNewFallsBackWithoutRegistryFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	registry, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := registry.Module("file-system"); !ok {
		t.Fatal("default registry not loaded")
	}
}

func TestParseModuleBitsListsLoadedState(t *testing.T) {
	registry := &Registry{Modules: map[string]Modules{
		"second": {ID: 2},
		"first":  {ID: 1},
		"third":  {ID: 3},
	}}

	got := registry.ParseModuleBits(1<<1 | 1<<3)
	want := "[*] first: loaded\n[*] second: unloaded\n[*] third: loaded"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
