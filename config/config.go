// Package config ...
package config

import (
	"time"
)

// Cfg global config holder
var Cfg Config

// Config config struct define
type Config struct {
	StatsRecord StatsRecordConfig `mapstructure:"stats-record"`
	Connection  struct {
		InfluxDB []InfluxClientConfig `mapstructure:"influxdb"`
		MySQL    []MySQLClientConfig  `mapstructure:"mysql"`
	} `mapstructure:"connection"`
	Points PointsConfig `mapstructure:"points"`
	Cases  CasesConfig  `mapstructure:"cases"`
}

// StatsRecordConfig stats record config
type StatsRecordConfig struct {
	Enable   bool   `mapstructure:"enable"`   // Record runtime statistics
	Host     string `mapstructure:"host"`     // Address of InfluxDB instance where runtime statistics will be recorded
	Database string `mapstructure:"database"` // Database that statistics will be written to
}

// InfluxClientConfig the influxdb client config struct
type InfluxClientConfig struct {
	Name          string `mapstructure:"name"`
	Default       bool   `mapstructure:"default"`
	URL           string `mapstructure:"url"`
	Precision     string `mapstructure:"precision"`
	Consistency   string `mapstructure:"consistency"`
	TLSSkipVerify bool   `mapstructure:"tls-skip-verify"`
	APIVersion    int    `mapstructure:"api-version"`
	V1            struct {
		Database        string `mapstructure:"db"`
		RetentionPolicy string `mapstructure:"rp"`
		User            string `mapstructure:"user"`
		Pass            string `mapstructure:"pass"`
	} `mapstructure:"v1"`
	V2 struct {
		OrgID  string `mapstructure:"org-id"`
		Bucket string `mapstructure:"bucket"`
		Token  string `mapstructure:"token"`
	} `mapstructure:"v2"`
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

// CasesConfig cases config
type CasesConfig struct {
	Delay       time.Duration `mapstructure:"delay"`
	Fast        bool          `mapstructure:"fast"`
	Tick        time.Duration `mapstructure:"tick"`
	CasesFile   string        `mapstructure:"cases-file"`
	CasesFilter []string      `mapstructure:"cases-filter"`
}
