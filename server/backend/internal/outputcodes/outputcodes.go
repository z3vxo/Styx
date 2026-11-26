package outputcodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Namespace int

const (
	NamespaceCore Namespace = iota
	NamespaceModule
	NamespaceWin32
)

type Table struct {
	core    map[int32]string
	win32   map[int32]string
	modules map[string]map[int32]string
	ranges  []moduleRange
}

type moduleRange struct {
	name  string
	start int32
	end   int32
}

type rawTable struct {
	Version int                  `json:"version"`
	Core    map[string]string    `json:"core"`
	Win32   map[string]string    `json:"win32"`
	Modules map[string]rawModule `json:"modules"`
}

type rawModule struct {
	Range string            `json:"range"`
	Codes map[string]string `json:"codes"`
}

func (t *Table) Message(namespace Namespace, code int32) (string, bool) {
	if t == nil {
		return "", false
	}
	switch namespace {
	case NamespaceCore:
		message, ok := t.core[code]
		return message, ok
	case NamespaceWin32:
		message, ok := t.win32[code]
		return message, ok
	case NamespaceModule:
		for _, entry := range t.ranges {
			if code >= entry.start && code <= entry.end {
				message, ok := t.modules[entry.name][code]
				return message, ok
			}
		}
	}
	return "", false
}

func Parse(data []byte) (*Table, error) {
	return parseTable(data)
}

func Load() (*Table, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}
	path := filepath.Join(home, ".local", "share", "styx", "modules", "codes.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read codes %s: %w", path, err)
	}
	table, err := parseTable(data)
	if err != nil {
		return nil, fmt.Errorf("parse codes %s: %w", path, err)
	}
	return table, nil
}

func parseTable(data []byte) (*Table, error) {
	var raw rawTable
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Version != 1 {
		return nil, fmt.Errorf("unsupported version %d", raw.Version)
	}

	core, err := parseMessages("core", raw.Core)
	if err != nil {
		return nil, err
	}
	win32, err := parseMessages("win32", raw.Win32)
	if err != nil {
		return nil, err
	}

	table := &Table{
		core:    core,
		win32:   win32,
		modules: make(map[string]map[int32]string, len(raw.Modules)),
	}
	for name, entry := range raw.Modules {
		start, end, err := parseRange(name, entry.Range)
		if err != nil {
			return nil, err
		}
		codes, err := parseMessages("module "+name, entry.Codes)
		if err != nil {
			return nil, err
		}
		if len(codes) == 0 {
			return nil, fmt.Errorf("module %q has no codes", name)
		}
		for code := range codes {
			if code < start || code > end {
				return nil, fmt.Errorf("module %q code %s is outside range %s", name, formatCode(code), entry.Range)
			}
		}
		table.modules[name] = codes
		table.ranges = append(table.ranges, moduleRange{name: name, start: start, end: end})
	}

	sort.Slice(table.ranges, func(i, j int) bool {
		if table.ranges[i].start != table.ranges[j].start {
			return table.ranges[i].start < table.ranges[j].start
		}
		return table.ranges[i].name < table.ranges[j].name
	})
	for index := 1; index < len(table.ranges); index++ {
		previous := table.ranges[index-1]
		current := table.ranges[index]
		if current.start <= previous.end {
			return nil, fmt.Errorf("module ranges %q and %q overlap", previous.name, current.name)
		}
	}

	return table, nil
}

func parseMessages(owner string, values map[string]string) (map[int32]string, error) {
	messages := make(map[int32]string, len(values))
	for rawCode, message := range values {
		code, err := parseCode(rawCode)
		if err != nil {
			return nil, fmt.Errorf("%s code %q: %w", owner, rawCode, err)
		}
		if _, exists := messages[code]; exists {
			return nil, fmt.Errorf("%s defines code %s more than once", owner, formatCode(code))
		}
		if strings.TrimSpace(message) == "" {
			return nil, fmt.Errorf("%s code %s has an empty message", owner, formatCode(code))
		}
		messages[code] = message
	}
	return messages, nil
}

func parseRange(owner, value string) (int32, int32, error) {
	parts := strings.SplitN(value, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("module %q range %q must look like \"200-299\"", owner, value)
	}
	start, err := parseCode(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("module %q range start: %w", owner, err)
	}
	end, err := parseCode(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("module %q range end: %w", owner, err)
	}
	if start > end {
		return 0, 0, fmt.Errorf("module %q range %q is backwards", owner, value)
	}
	return start, end, nil
}

func parseCode(value string) (int32, error) {
	code, err := strconv.ParseInt(strings.TrimSpace(value), 0, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid code %q", value)
	}
	return int32(code), nil
}

func formatCode(code int32) string {
	return fmt.Sprintf("%d (0x%X)", code, uint32(code))
}
