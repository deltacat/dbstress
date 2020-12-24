package runner

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/report"
	"github.com/deltacat/dbstress/stress"
	"github.com/sirupsen/logrus"
)

var (
	tick                       time.Duration
	fast, quiet, kapacitorMode bool
	pointsCfg                  config.PointsConfig
	pointsN                    uint64
)

// Runner runner interface
type Runner interface {
	Run() error
}

type caseRunner struct {
	cli         client.Client
	cfg         config.CaseConfig
	concurrency int
}

// Setup runner context
func Setup(_tick time.Duration, _fast, _quiet, _kapacitorMode bool, ptsCfg config.PointsConfig) {
	tick = _tick
	fast = _fast
	quiet = _quiet
	kapacitorMode = _kapacitorMode
	pointsCfg = ptsCfg
	if pointsCfg.PointsN == 0 {
		pointsN = math.MaxUint64
	} else {
		pointsN = pointsCfg.PointsN
	}
}

func (r *caseRunner) doInsert(doWrite doWriteFunc) error {

	sink := stress.NewMultiSink(r.concurrency)
	sink.AddSink(stress.NewErrorSink(r.concurrency))

	// todo
	// if recordStats {
	// 	sink.AddSink(stress.NewInfluxDBSink(int(concurrency), statsHost, statsDB))
	// }

	sink.Open()

	var wg sync.WaitGroup
	wg.Add(r.concurrency)

	start := time.Now()

	totalWritten, err := doWrite(sink.Chan())

	totalTime := time.Since(start)
	if err := r.cli.Close(); err != nil {
		logrus.WithError(err).Error("Error closing client")
	}

	sink.Close()
	throughput := int(float64(totalWritten) / totalTime.Seconds())
	if quiet {
		fmt.Println(throughput)
	} else {

		report.Append([]string{
			r.cli.Name(),
			r.cfg.Connection,
			"insert",
			fmt.Sprintf("%d", r.concurrency),
			fmt.Sprintf("%d", r.cfg.BatchSize),
			fmt.Sprintf("%.3fs", totalTime.Seconds()),
			fmt.Sprintf("%d", throughput),
			fmt.Sprintf("%d", totalWritten)})

		logrus.WithField("Write Throughput:", throughput).WithField("Points Written:", totalWritten).Info("run stress done")
	}

	return err
}
