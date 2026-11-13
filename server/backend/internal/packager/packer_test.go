package packager

import (
	"encoding/binary"
	"testing"
)

func TestPackTasksWritesLoadModule(t *testing.T) {
	moduleData := []byte{1, 2, 3}
	out := PackTasks([]TaskRecord{{TaskID: 9, Kind: "load-module", ModuleID: 1, ModuleData: moduleData}})

	values := []int32{
		int32(binary.LittleEndian.Uint32(out[0:])),
		int32(binary.LittleEndian.Uint32(out[4:])),
		int32(binary.LittleEndian.Uint32(out[8:])),
		int32(binary.LittleEndian.Uint32(out[12:])),
		int32(binary.LittleEndian.Uint32(out[16:])),
	}
	want := []int32{1, 9, 0xb1, 1, int32(len(moduleData))}
	if len(out) != 20+len(moduleData) {
		t.Fatalf("length = %d", len(out))
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("value %d = %d, want %d", index, values[index], want[index])
		}
	}
	if string(out[20:]) != string(moduleData) {
		t.Fatalf("module data = %v", out[20:])
	}
}

func TestPackTasksWritesRunCommandArgs(t *testing.T) {
	args := `[{"type":"int32","value":7},{"type":"string","value":"path"}]`
	out := PackTasks([]TaskRecord{{TaskID: 4, Kind: "run-command", ModuleID: 1, Code: 4, Args: args}})

	values := []int32{
		int32(binary.LittleEndian.Uint32(out[0:])),
		int32(binary.LittleEndian.Uint32(out[4:])),
		int32(binary.LittleEndian.Uint32(out[8:])),
		int32(binary.LittleEndian.Uint32(out[12:])),
		int32(binary.LittleEndian.Uint32(out[16:])),
		int32(binary.LittleEndian.Uint32(out[20:])),
		int32(binary.LittleEndian.Uint32(out[24:])),
		int32(binary.LittleEndian.Uint32(out[28:])),
	}
	want := []int32{1, 4, 0xb3, 1, 4, 2, 7, 4}
	if len(out) != 36 {
		t.Fatalf("length = %d", len(out))
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("value %d = %d, want %d", index, values[index], want[index])
		}
	}
	if string(out[32:]) != "path" {
		t.Fatalf("string arg = %q", out[32:])
	}
}
