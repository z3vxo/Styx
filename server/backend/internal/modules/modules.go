package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Registry struct {
	Modules map[string]Modules
	Data    map[string][]byte
}

type Modules struct {
	ID       uint8             `json:"id"`
	File     string            `json:"file"`
	Commands map[string]uint32 `json:"commands"`
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
		if code, ok := module.Commands[name]; ok {
			return moduleName, module, code, true
		}
	}
	return "", Modules{}, 0, false
}

func Default() *Registry {
	return &Registry{Modules: map[string]Modules{
		"file-system": {
			ID:       1,
			File:     "file-system.mod",
			Commands: map[string]uint32{"ls": 1, "cp": 2, "pwd": 3, "cat": 4, "mv": 5},
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
