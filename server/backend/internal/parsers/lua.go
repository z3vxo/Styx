package parsers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	lua "github.com/yuin/gopher-lua"
)

const maxReadBytes = 64 << 20

func newReaderMetatable(L *lua.LState) *lua.LTable {
	reader := L.NewTypeMetatable("reader")
	L.SetField(reader, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"read4": func(L *lua.LState) int {
			r := L.CheckUserData(1).Value.(*bytes.Reader)
			var buf [4]byte
			if _, err := io.ReadFull(r, buf[:]); err != nil {
				L.RaiseError("read4: %v", err)
				return 0
			}
			L.Push(lua.LNumber(binary.LittleEndian.Uint32(buf[:])))
			return 1
		},
		"read8": func(L *lua.LState) int {
			r := L.CheckUserData(1).Value.(*bytes.Reader)
			var buf [8]byte
			if _, err := io.ReadFull(r, buf[:]); err != nil {
				L.RaiseError("read8: %v", err)
				return 0
			}
			L.Push(lua.LNumber(binary.LittleEndian.Uint64(buf[:])))
			return 1
		},
		"readBytes": func(L *lua.LState) int {
			r := L.CheckUserData(1).Value.(*bytes.Reader)
			n := L.CheckInt(2)
			if n < 0 {
				L.ArgError(2, "must not be negative")
				return 0
			}
			if n > maxReadBytes {
				L.ArgError(2, "length exceeds limit")
				return 0
			}
			buf := make([]byte, n)
			if _, err := io.ReadFull(r, buf); err != nil {
				L.RaiseError("readBytes: %v", err)
				return 0
			}
			L.Push(lua.LString(buf))
			return 1
		},
	}))
	return reader
}

func (p *Parser) LoadLuaScript(parseID int32, scriptPath string) error {
	chunk, err := p.L.LoadFile(scriptPath)
	if err != nil {
		return err
	}
	if err := p.L.CallByParam(lua.P{Fn: chunk, NRet: 1, Protect: true}); err != nil {
		return err
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	fn, ok := ret.(*lua.LFunction)
	if !ok {
		return fmt.Errorf("%s must return a function", scriptPath)
	}
	p.parsers[parseID] = fn
	return nil
}

func (p *Parser) Parse(parseID int32, data *bytes.Reader) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fn, ok := p.parsers[parseID]
	if !ok {
		return "", fmt.Errorf("unknown parser id %d", parseID)
	}

	ud := p.L.NewUserData()
	ud.Value = data
	p.L.SetMetatable(ud, p.mt)

	if err := p.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, ud); err != nil {
		return "", err
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	output, ok := ret.(lua.LString)
	if !ok {
		return "", fmt.Errorf("parser %d returned %s, want string", parseID, ret.Type())
	}
	return string(output), nil
}
