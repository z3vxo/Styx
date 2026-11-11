package packager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type AgentRegister struct {
	AgentID  int32
	Pid      int32
	PPid     int32
	Arch     string
	IsAdmin  int32
	Username string
	Hostname string
	Domain   string
	Country  string
}

type Reader struct {
	r   *bytes.Reader
	err error
}

func (r *Reader) unPack4() int32 {
	if r.err != nil {
		return 0
	}
	var val int32
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func (r *Reader) unPackBytes(n int32) []byte {
	if r.err != nil {
		return nil
	}
	buf := make([]byte, n)
	_, r.err = io.ReadFull(r.r, buf)
	return buf
}

func (r *Reader) unPack8() int64 {
	if r.err != nil {
		return 0
	}
	var val int64
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func ParseIncomingMetaDataAndValidate(body *bytes.Reader) (int32, int32, error) {
	r := &Reader{r: body}

	Magic := r.unPackBytes(4)
	AgentID := r.unPack4()
	MsgID := r.unPack4()

	if r.err != nil {
		return 0, 0, r.err
	}
	if !bytes.Equal(Magic, []byte{0xff, 0xf1, 0xf2, 0xf3}) {
		fmt.Println("[!] MetaData doesnt match!")
		return 0, 0, fmt.Errorf("incorrect magic value")
	}

	return MsgID, AgentID, r.err
}

/*
 	[AgentID]      4 Bytes
  	[PID]          4 Bytes
    [PPID]         4 Bytes
    [Arch]    	   4 Bytes
    [IsAdmin]      4 Bytes
    [Username Len] 4 Bytes
    [Username Str] N Bytes
    [Hostname Len] 4 Bytes
    [Hostname Str] N Bytes
    [Domain Len]   4 Bytes
    [Domain Str]   N Bytes
    [Country Len]  4 Bytes
    [Country Str]  4 Bytes
*/

func ParseRegister(body *bytes.Reader) (AgentRegister, error) {
	r := &Reader{r: body}

	Agent := AgentRegister{}

	Agent.AgentID = r.unPack4()
	Agent.Pid = r.unPack4()
	Agent.PPid = r.unPack4()
	arch := r.unPack4()
	if arch == 1 {
		Agent.Arch = "x64"
	} else {
		Agent.Arch = "x86"
	}
	Agent.IsAdmin = r.unPack4()
	UserLen := r.unPack4()
	Agent.Username = string(r.unPackBytes(UserLen))
	HostLen := r.unPack4()
	Agent.Hostname = string(r.unPackBytes(HostLen))
	DomainLen := r.unPack4()
	Agent.Domain = string(r.unPackBytes(DomainLen))
	CountryLen := r.unPack4()
	Agent.Country = string(r.unPackBytes(CountryLen))

	return Agent, r.err
}
