package client

import (
	"github.com/deltacat/dbstress/config"
)

// Client db connection client interface
type Client interface {
	Create(cmd string) error
	Send([]byte) (latNs int64, statusCode int, body string, err error)
	SendString(query string) (latNs int64, statusCode int, body string, err error)
	Close() error
	Reset() error
	Name() string
	GzipLevel() int // temp function for refactoring
}

// InfluxConfig influxdb client config
type InfluxConfig = config.InfluxClientConfig

// MySQLConfig mysql client config
type MySQLConfig = config.MySQLClientConfig
