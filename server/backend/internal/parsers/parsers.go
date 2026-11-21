package parser

import (
	"github.com/yuin/gopher-lua"
	"sync"
)

type Parser struct {
	L *lua.LState
	parsers map[int32]*lua.LFunction
	mu sync.Mutex
}

func New() (*Parsers, error) {
	l, err := SetupLuaEnv()
	if err != nil {
		return nil, err
	}

	return &Parser{L: l, parser: make(map[int32]*lua.LFunction)}, nil
}