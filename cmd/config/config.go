// Package config ...
package config

// Cfg global config holder
var Cfg Config

// Config config struct define
type Config struct {
	Connection struct {
		InfluxDB string `mapstructure:"influxdb"`
		Mysql    string `mapstructure:"mysql"`
	} `mapstructure:"connection"`
}
