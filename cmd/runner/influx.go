package runner

import (
	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
)

// InfluxRunner influxdb runner
type InfluxRunner struct {
	caseRunner
}

// NewInfluxRunner create a new mysql runner instance
func NewInfluxRunner(cli client.Client, cs config.CaseConfig) InfluxRunner {
	return InfluxRunner{
		caseRunner: caseRunner{
			cli:         cli,
			cfg:         cs,
			concurrency: cs.Concurrent,
		},
	}
}

// Run run the case
func (r *InfluxRunner) Run() error {
	defer r.cli.Close()
	if !kapacitorMode {
		if err := r.cli.Create(""); err != nil {
			return err
		}
	}

	return r.doInsert(r.doWriteInflux)
}
