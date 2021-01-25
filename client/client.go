package client

import (
	"github.com/deltacat/dbstress/config"
)

// Client db connection client interface
type Client interface {
	Create(cmd string) error
	Send(b []byte, gzip int) (latNs int64, statusCode int, body string, err error)
	SendString(query string) (latNs int64, statusCode int, body string, err error)
	Close() error
	Reset() error
	Name() string
	Connection() string // return connection to check
}

// InfluxConfig influxdb client config
type InfluxConfig = config.InfluxClientConfig

// MySQLConfig mysql client config
type MySQLConfig = config.MySQLClientConfig
