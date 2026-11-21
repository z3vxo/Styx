package parsers

import (
	"github.com/yuin/gopher-lua"
	"os"
	"filepath"
)


func SetupLuaEnv() (*lua.LState, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	moduleDir := filepath.Join(home, ".local", "styx", "modules")

	L := lua.NewState()

	err := filepath.WalkDir(moduleDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || filepath.Ext(path) != ".lua" {
			return nil
		}

		rel, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(rel, ".lua")
		name = strings.ReplaceAll(name, string(filepath.Separator), ".")

		fn, err := L.LoadFile(path)
		if err != nil {
			return err
		}

		L.PreloadModule(name, func(state *lua.LState) int {
			state.Push(fn)
			state.Call(0, 1)
		})  
		return nil
	})

	if err != nil {
		L.Close()
		return nil, err
	}

	return L, nil


}