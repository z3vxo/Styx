package agentsrv

const (
	MSG_GET_TASK = 0x1
	MSG_REGISTER = 0x2
	MSG_OUTPUT   = 0x3

	PARSE_MODULE_BITS = 101
)

type ErrorMsg struct {
	Msg string `json:"error_msg"`
}
