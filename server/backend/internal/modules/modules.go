package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Registry struct {
	Modules map[string]Modules
	Data    map[string][]byte
}

type Modules struct {
	ABI         int                `json:"abi,omitempty"`
	ID          uint8              `json:"id"`
	File        string             `json:"file"`
	Description string             `json:"description,omitempty"`
	Commands    map[string]Command `json:"commands"`
}

type Command struct {
	Code        uint32       `json:"code"`
	Usage       string       `json:"usage,omitempty"`
	Description string       `json:"description,omitempty"`
	Args        []CommandArg `json:"args,omitempty"`
}

type CommandArg struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Default     any    `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var code uint32
	if err := json.Unmarshal(data, &code); err == nil {
		c.Code = code
		return nil
	}

	type commandJSON Command
	var parsed commandJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	if parsed.Code == 0 {
		return fmt.Errorf("command code is required")
	}
	if parsed.Args == nil {
		parsed.Args = []CommandArg{}
	}
	*c = Command(parsed)
	return nil
}

func (mod *Registry) IsValidModuleName(name string) bool {
	_, ok := mod.Modules[name]
	return ok
}

func (mod *Registry) Module(name string) (Modules, bool) {
	module, ok := mod.Modules[name]
	return module, ok
}

func (mod *Registry) ModuleData(name string) []byte {
	return mod.Data[name]
}

func (mod *Registry) ResolveCommand(name string) (string, Modules, uint32, bool) {
	for moduleName, module := range mod.Modules {
		if command, ok := module.Commands[name]; ok {
			return moduleName, module, command.Code, true
		}
	}
	return "", Modules{}, 0, false
}

// ParseModuleBits renders the agent's loaded-module bitmap as one line per
// registered module. Bit N is set when module ID N is loaded.
func (mod *Registry) ParseModuleBits(bits int64) string {
	names := make([]string, 0, len(mod.Modules))
	for name := range mod.Modules {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		left, right := mod.Modules[names[i]], mod.Modules[names[j]]
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		return names[i] < names[j]
	})

	lines := make([]string, 0, len(names))
	for _, name := range names {
		module := mod.Modules[name]
		state := "unloaded"
		if module.ID < 64 && (uint64(bits)>>module.ID)&1 == 1 {
			state = "loaded"
		}
		lines = append(lines, fmt.Sprintf("[*] %s: %s", name, state))
	}
	return strings.Join(lines, "\n")
}

func Default() *Registry {
	return &Registry{Modules: map[string]Modules{
		"file-system": {
			ABI:         1,
			ID:          1,
			File:        "file-system.mod",
			Description: "Local filesystem operations",
			Commands: map[string]Command{
				"ls": {
					Code:  1,
					Usage: "ls [path]",
					Args: []CommandArg{
						{Name: "path", Type: "string", Default: "."},
					},
				},
				"cp": {
					Code:  2,
					Usage: "cp <src> <dst>",
					Args: []CommandArg{
						{Name: "src", Type: "string", Required: true},
						{Name: "dst", Type: "string", Required: true},
					},
				},
				"pwd": {Code: 3, Usage: "pwd", Args: []CommandArg{}},
				"cat": {
					Code:  4,
					Usage: "cat <path>",
					Args: []CommandArg{
						{Name: "path", Type: "string", Required: true},
					},
				},
				"mv": {
					Code:  5,
					Usage: "mv <src> <dst>",
					Args: []CommandArg{
						{Name: "src", Type: "string", Required: true},
						{Name: "dst", Type: "string", Required: true},
					},
				},
				"cd": {
					Code:  6,
					Usage: "cd <path>",
					Args: []CommandArg{
						{Name: "path", Type: "string", Required: true},
					},
				},
			},
		},
	}}
}

func New() (*Registry, error) {
	registry := Default()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".local", "share", "styx", "modules", "modules.json")

	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("open module registry %s: %w", path, err)
		}
	} else {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&registry.Modules); err != nil {
			return nil, fmt.Errorf("decode module registry %s: %w", path, err)
		}
	}

	if err := registry.loadData(); err != nil {
		return nil, err
	}
	return registry, nil
}

func (mod *Registry) loadData() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pattern := filepath.Join(home, ".local", "share", "styx", "modules", "*", "*.mod")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	data := make(map[string][]byte)
	for _, path := range matches {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.Base(filepath.Dir(path))
		data[name] = contents
	}
	mod.Data = data
	return nil
}
