package runner

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/data/influx/lineprotocol"
	"github.com/deltacat/dbstress/data/influx/point"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/deltacat/dbstress/stress"
)

type doWriteFunc func(resultChan chan stress.WriteResult) (uint64, error)

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
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: uint64(r.cfg.BatchSize),
				MaxPoints: pointsN / uint64(r.concurrency), // divide by concurreny
				GzipLevel: r.cli.GzipLevel(),
				Deadline:  time.Now().Add(r.cfg.Runtime),
				Tick:      tick,
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
