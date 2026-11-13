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

var taskWireCodes = map[string]int32{
	"load-module":   0xb1,
	"list-modules":  0xb2,
	"run-command":   0xb3,
	"unload-module": 0xb4,
}

func PackTasks(tasks []TaskRecord) []byte {
	w := &Writer{}
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

func writeString(w *Writer, value string) {
	w.Pack4(int32(len(value)))
	w.PackBytes([]byte(value))
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
