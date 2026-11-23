package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"styx/internal/modules"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type Parser struct {
	L       *lua.LState
	mt      *lua.LTable
	parsers map[int32]*lua.LFunction
	mu      sync.Mutex
}

func New() (*Parser, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	modulesPath := filepath.Join(home, ".local", "share", "styx", "modules")

	f, err := os.Open(filepath.Join(modulesPath, "modules.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reg modules.Registry
	err = json.NewDecoder(f).Decode(&reg.Modules)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		L:       lua.NewState(),
		parsers: make(map[int32]*lua.LFunction),
	}
	p.mt = newReaderMetatable(p.L)

	for moduleName, module := range reg.Modules {
		for cmdName, cmd := range module.Commands {
			if cmd.ParseID == 0 {
				continue
			}
			script := filepath.Join(modulesPath, moduleName, cmdName, cmdName+".lua")
			if err := p.LoadLuaScript(cmd.ParseID, script); err != nil {
				return nil, fmt.Errorf("load parser %s.%s: %w", moduleName, cmdName, err)
			}
		}
	}

	return p, nil

}
