// Package config ...
package config

// Cfg global config holder
var Cfg Config

// Config config struct define
type Config struct {
	Connection struct {
		Influxdb InfluxdbClientConfig `mapstructure:"influxdb"`
		Mysql    MySQLClientConfig    `mapstructure:"mysql"`
	} `mapstructure:"connection"`
	Points PointsConfig `mapstructure:"points"`
	Tests  []struct {
		Name       string `mapstructure:"name"`
		Type       string `mapstructure:"type"`
		Connection string `mapstructure:"connection"`
		Case       string `mapstructure:"case"`
	} `mapstructure:"tests"`
}

// InfluxdbClientConfig the influxdb client config struct
type InfluxdbClientConfig struct {
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
}
