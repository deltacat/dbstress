package runner

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/data/mysql"
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
	Info() map[string]interface{}
}

type caseRunner struct {
	cli         client.Client
	cfg         config.CaseConfig
	concurrency int
}

type doWriteFunc func(resultChan chan stress.WriteResult) (uint64, error)

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
	report.SetHeader([]string{"case", "connection", "action", "concur", "batch", "time", "run", "throughput", "points"})
}

// Close finish all runners
func Close() {
	if !quiet {
		fmt.Printf("\nReport: =======>\n")
		fmt.Printf("Use point template: %s %s <timestamp>\n\n", pointsCfg.SeriesKey, pointsCfg.FieldsStr)
		report.Render()
		fmt.Println()
	}
}

// BuildAllRunners build runner from cases config
func BuildAllRunners(cfg config.Config) []Runner {
	cfs := cfg.Cases.Cases
	runners := []Runner{}
	for _, cf := range cfs {
		if strings.Contains(strings.ToLower(cf.Name), "influx") {
			if cof, err := cfg.FindInfluxDBConnection(cf.Connection); err == nil {
				if cli, err := client.NewInfluxClient(cof, ""); err == nil {
					r := NewInfluxRunner(cli, cf)
					runners = append(runners, &r)
				}
			}
		} else if strings.Contains(strings.ToLower(cf.Name), "mysql") {
			if cof, err := cfg.FindMySQLConnection(cf.Connection); err == nil {
				if cli, err := client.NewMySQLClient(cof); err == nil {
					if layout, err := mysql.GenerateLayout(pointsCfg.Measurement, pointsCfg.SeriesKey, pointsCfg.FieldsStr); err == nil {
						r := NewMySQLRunner(cli, cf, layout)
						runners = append(runners, &r)
					}
				}
			}
		}
	}
	return runners
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
			r.cfg.Name,
			r.cli.Connection(),
			"insert",
			fmt.Sprintf("%d", r.concurrency),
			fmt.Sprintf("%d", r.cfg.BatchSize),
			fmt.Sprintf("%.0fs", r.cfg.Runtime.Seconds()),
			fmt.Sprintf("%.3fs", totalTime.Seconds()),
			fmt.Sprintf("%d", throughput),
			fmt.Sprintf("%d", totalWritten)})
	}

	return err
}

func (r *caseRunner) Info() map[string]interface{} {
	return map[string]interface{}{
		"name":       r.cfg.Name,
		"connection": r.cli.Connection(),
	}
}
