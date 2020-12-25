package runner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/deltacat/dbstress/stress"
)

// MySQLRunner mysql runner
type MySQLRunner struct {
	caseRunner
	layout mysql.Layout
}

// NewMySQLRunner create a new mysql runner instance
func NewMySQLRunner(cli client.Client, cs config.CaseConfig, layout mysql.Layout) MySQLRunner {
	return MySQLRunner{
		caseRunner: caseRunner{
			cli:         cli,
			cfg:         cs,
			concurrency: cs.Concurrent,
		},
		layout: layout,
	}
}

// Run run the case
func (r *MySQLRunner) Run() error {
	if err := r.cli.Create(r.layout.GetCreateStmt()); err != nil {
		return err
	}

	return r.doInsert(r.doWriteMysql)
}

func (r *MySQLRunner) doWriteMysql(resultChan chan stress.WriteResult) (uint64, error) {

	var wg sync.WaitGroup
	wg.Add(int(r.concurrency))

	seriesN := pointsCfg.SeriesN

	totalWritten := uint64(0)
	startSplit := 0
	inc := int(seriesN) / int(r.concurrency)
	endSplit := inc

	tbl := mysql.NewTableChunk(r.layout, uint64(r.cfg.BatchSize))
	for i := uint64(0); i < uint64(r.concurrency); i++ {

		go func(startSplit, endSplit int) {
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: uint64(r.cfg.BatchSize),
				MaxPoints: pointsN / uint64(r.concurrency), // divide by concurreny
				Deadline:  time.Now().Add(r.cfg.Runtime),
				Tick:      tick,
				Results:   resultChan,
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.WriteMySQL(tbl, r.cli, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()

	return totalWritten, nil
}
