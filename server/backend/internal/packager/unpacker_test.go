package packager

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"styx/internal/modules"
	"styx/internal/parsers"
)

func newTestParser(t *testing.T) *parsers.Parser {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	modulesPath := filepath.Join(home, ".local", "share", "styx", "modules")
	commandPath := filepath.Join(modulesPath, "test", "echo")
	if err := os.MkdirAll(commandPath, 0755); err != nil {
		t.Fatal(err)
	}

	const registry = `{"test":{"id":102,"file":"test.mod","commands":{"echo":{"code":1,"parse_id":102}}}}`
	if err := os.WriteFile(filepath.Join(modulesPath, "modules.json"), []byte(registry), 0644); err != nil {
		t.Fatal(err)
	}

	const script = `return function(r) return "value=" .. r:read4() end`
	if err := os.WriteFile(filepath.Join(commandPath, "echo.lua"), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := parsers.New()
	if err != nil {
		t.Fatal(err)
	}
	return parser
}

func TestParseAgentOutputUsesRegisteredParser(t *testing.T) {
	parser := newTestParser(t)

	w := &Writer{}
	w.Pack4(1)
	w.Pack4(4242)
	w.Pack4(OUTPUT_NEEDS_PARSE)
	w.Pack4(102)
	w.Pack4(42)

	out, err := ParseAgentOutput(bytes.NewReader(w.Bytes()), 7, nil, parser)
	if err != nil {
		t.Fatal(err)
	}
	if out.Total != 1 || len(out.Outputs) != 1 {
		t.Fatalf("outputs = %+v", out.Outputs)
	}
	if out.Outputs[0].TaskID != 4242 {
		t.Fatalf("task id = %d", out.Outputs[0].TaskID)
	}
	if out.Outputs[0].Output != "value=42" {
		t.Fatalf("output = %q", out.Outputs[0].Output)
	}
}

func TestParseAgentOutputKeepsModuleBits(t *testing.T) {
	parser := newTestParser(t)
	registry := &modules.Registry{Modules: map[string]modules.Modules{
		"first":  {ID: 1},
		"second": {ID: 2},
	}}

	w := &Writer{}
	w.Pack4(1)
	w.Pack4(1)
	w.Pack4(OUTPUT_NEEDS_PARSE)
	w.Pack4(PARSE_MODULE_BITS)
	w.Pack8(1 << 1)

	out, err := ParseAgentOutput(bytes.NewReader(w.Bytes()), 7, registry, parser)
	if err != nil {
		t.Fatal(err)
	}
	want := "[*] first: loaded\n[*] second: unloaded"
	if out.Outputs[0].Output != want {
		t.Fatalf("output = %q, want %q", out.Outputs[0].Output, want)
	}
}
