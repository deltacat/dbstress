package client

// Client db connection client interface
type Client interface {
	Create(string) error
	Send([]byte) (latNs int64, statusCode int, body string, err error)
	Close() error
}
