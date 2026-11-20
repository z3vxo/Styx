package packager

import (
	"encoding/binary"
	"testing"
)

func TestPackTasksWithoutTasksReturnsNoTasks(t *testing.T) {
	for _, tasks := range [][]TaskRecord{nil, {}} {
		out := PackTasks(tasks)
		if len(out) != 4 {
			t.Fatalf("length = %d, want 4", len(out))
		}
		if got := int32(binary.LittleEndian.Uint32(out)); got != RESP_NO_TASKS {
			t.Fatalf("response = %#x, want %#x", got, RESP_NO_TASKS)
		}
	}
}

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
	want := []int32{1, 0xb1, 9, 1, int32(len(moduleData))}
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

func TestPackTasksWritesUnloadModule(t *testing.T) {
	out := PackTasks([]TaskRecord{{TaskID: 5, Kind: "unload-module", ModuleID: 2}})

	if len(out) != 16 {
		t.Fatalf("length = %d", len(out))
	}
	values := []int32{
		int32(binary.LittleEndian.Uint32(out[0:])),
		int32(binary.LittleEndian.Uint32(out[4:])),
		int32(binary.LittleEndian.Uint32(out[8:])),
		int32(binary.LittleEndian.Uint32(out[12:])),
	}
	want := []int32{1, 0xb4, 5, 2}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("value %d = %d, want %d", index, values[index], want[index])
		}
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
	}
	want := []int32{1, 0xb3, 4, 1, 4, 7, 5}
	if len(out) != 33 {
		t.Fatalf("length = %d", len(out))
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("value %d = %d, want %d", index, values[index], want[index])
		}
	}
	if string(out[28:]) != "path\x00" {
		t.Fatalf("string arg = %q", out[28:])
	}
}
