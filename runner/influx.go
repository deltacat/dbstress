package runner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/data/influx/lineprotocol"
	"github.com/deltacat/dbstress/data/influx/point"
	"github.com/deltacat/dbstress/stress"
)

// InfluxRunner influxdb runner
type InfluxRunner struct {
	caseRunner
}

// NewInfluxRunner create a new mysql runner instance
func NewInfluxRunner(cli client.Client, cs CaseConfig) InfluxRunner {
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

func (r *InfluxRunner) doWriteInflux(resultChan chan stress.WriteResult) (uint64, error) {
	var wg sync.WaitGroup
	wg.Add(r.concurrency)

	seriesN := pointsCfg.SeriesN

	var totalWritten uint64
	startSplit := 0
	inc := int(seriesN) / int(r.concurrency)
	endSplit := inc

	pts := point.NewPoints(pointsCfg.Measurement, pointsCfg.SeriesKey, pointsCfg.FieldsStr, seriesN, lineprotocol.Nanosecond)
	for i := uint64(0); i < uint64(r.concurrency); i++ {

		go func(startSplit, endSplit int) {
			cfg := stress.WriteConfig{
				BatchSize: uint64(r.cfg.BatchSize),
				MaxPoints: pointsN / uint64(r.concurrency), // divide by concurreny
				GzipLevel: r.cli.GzipLevel(),
				Deadline:  time.Now().Add(r.cfg.Runtime.Duration),
				Tick:      time.Tick(tick),
				Results:   resultChan,
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.WriteInflux(pts[startSplit:endSplit], r.cli, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()

	return totalWritten, nil
}
