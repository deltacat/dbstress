package runner

import "github.com/deltacat/dbstress/csv"

// CaseConfig test case config
type CaseConfig struct {
	Name       string       `mapstructure:"name"`
	Connection string       `mapstructure:"connection"`
	Concurrent int          `mapstructure:"concurrent"`
	BatchSize  int          `mapstructure:"batch-size"`
	Runtime    csv.Duration `mapstructure:"runtime"`
}
