package builder

type Builder struct {
	BackendURL  string
	BackendPort int
}

func NewBuilder(host string, port int) *Builder {
	return &Builder{BackendURL: host, BackendPort: port}
}
