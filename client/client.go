package client

import "github.com/deltacat/dbstress/config"

// Client db connection client interface
type Client interface {
	Create(string) error
	Send([]byte) (latNs int64, statusCode int, body string, err error)
	Close() error
}

// InfluxConfig influxdb client config
type InfluxConfig = config.InfluxdbClientConfig

// MySQLConfig mysql client config
type MySQLConfig = config.MySQLClientConfig
