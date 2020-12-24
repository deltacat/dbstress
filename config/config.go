// Package config ...
package config

import "time"

// Cfg global config holder
var Cfg Config

// Config config struct define
type Config struct {
	Connection struct {
		InfluxDB []InfluxClientConfig `mapstructure:"influxdb"`
		MySQL    []MySQLClientConfig  `mapstructure:"mysql"`
	} `mapstructure:"connection"`
	Points PointsConfig `mapstructure:"points"`
	Cases  struct {
		Delay time.Duration `mapstructure:"delay"`
		Cases []CaseConfig  `mapstructure:"case"`
	} `mapstructure:"cases"`
}

// InfluxClientConfig the influxdb client config struct
type InfluxClientConfig struct {
	Name            string `mapstructure:"name"`
	Default         bool   `mapstructure:"default"`
	URL             string `mapstructure:"url"`
	Database        string `mapstructure:"db"`
	RetentionPolicy string `mapstructure:"rp"`
	User            string `mapstructure:"user"`
	Pass            string `mapstructure:"pass"`
	Precision       string `mapstructure:"precision"`
	Consistency     string `mapstructure:"consistency"`
	TLSSkipVerify   bool   `mapstructure:"tls-skip-verify"`
	Gzip            int    `mapstructure:"gzip"`
}

// MySQLClientConfig mysql client config
type MySQLClientConfig struct {
	Name     string `mapstructure:"name"`
	Default  bool   `mapstructure:"default"`
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Pass     string `mapstructure:"pass"`
	Database string `mapstructure:"db"`
}

// PointsConfig points to write config
type PointsConfig struct {
	Measurement string `mapstructure:"measurement"`
	SeriesKey   string `mapstructure:"series-key"`
	FieldsStr   string `mapstructure:"fields-str"`
	SeriesN     int    `mapstructure:"series-num"`
	PointsN     uint64 `mapstructure:"points-num"`
}

// CaseConfig test case config
type CaseConfig struct {
	Name       string        `mapstructure:"name"`
	Connection string        `mapstructure:"connection"`
	Concurrent int           `mapstructure:"concurrent"`
	BatchSize  int           `mapstructure:"batch-size"`
	Runtime    time.Duration `mapstructure:"runtime"`
}
