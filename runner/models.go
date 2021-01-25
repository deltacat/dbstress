package runner

import "github.com/deltacat/dbstress/csv"

// CaseConfig test case config
type CaseConfig struct {
	Name       string       `mapstructure:"name"`
	Connection string       `mapstructure:"connection"`
	Concurrent int          `mapstructure:"concurrent"`
	BatchSize  int          `mapstructure:"batch-size"`
	Gzip       int          `mapstructure:"gzip"` // If non-zero, gzip write bodies with given compression level. 1=best speed, 9=best compression, -1=gzip default.
	Runtime    csv.Duration `mapstructure:"runtime"`
}
