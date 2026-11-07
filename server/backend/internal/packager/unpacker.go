package packager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

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

func ParseIncomingMetaDataAndValidate(body *bytes.Reader) (int32, error) {
	r := &Reader{r: body}

	Magic := r.unPackBytes(4)
	AgentID := r.unPack4()

	if !bytes.Equal(Magic, []byte{0xff, 0xf1, 0xf2, 0xf3}) {
		return 0, fmt.Errorf("incorrect magic value")
	}

	return AgentID, r.err
}
