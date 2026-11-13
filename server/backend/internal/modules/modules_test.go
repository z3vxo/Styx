package modules

import "testing"

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
