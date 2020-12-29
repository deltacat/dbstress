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
	"github.com/deltacat/dbstress/utils"
	"github.com/sirupsen/logrus"
)

var (
	tick                       time.Duration
	fast, quiet, kapacitorMode bool
	pointsCfg                  config.PointsConfig
	pointsN                    uint64
	statsHost, statsDB         string
	recordStats                bool
)

// Runner runner interface
type Runner interface {
	Run() error

	// Info return a map to print log
	Info() map[string]interface{}
	// Result return a map to print log
	Result() map[string]interface{}
}

type caseRunner struct {
	cli client.Client
	cfg config.CaseConfig

	concurrency  int
	totalTime    time.Duration
	totalWritten uint64
	throughput   uint64
}

type doWriteFunc func(resultChan chan stress.WriteResult) (uint64, error)

// Setup runner context
func Setup(_tick time.Duration, _fast, _quiet, _kapacitorMode bool, ptsCfg config.PointsConfig, statsCfg config.StatsRecordConfig) {
	tick = _tick
	fast = _fast
	quiet = _quiet
	kapacitorMode = _kapacitorMode
	recordStats = statsCfg.Enable
	statsHost = statsCfg.Host
	statsDB = statsCfg.Database
	pointsCfg = ptsCfg
	if pointsCfg.PointsN == 0 {
		pointsN = math.MaxUint64
	} else {
		pointsN = pointsCfg.PointsN
	}
	report.SetHeader([]string{"case", "connection", "action", "concur", "batch", "start", "run", "throughput", "points"})
}

// Close finish all runners
func Close() {
}

// Report print report
func Report() {
	if !quiet {
		fmt.Printf("\nReport: =======>\n")
		fmt.Printf("Use point template: %s %s <timestamp>\n\n", pointsCfg.SeriesKey, pointsCfg.FieldsStr)
		report.Render()
		fmt.Println()
	}
}

// BuildAllRunners build runner from cases config
func BuildAllRunners(cfg config.Config, filters []string) []Runner {
	cfs := cfg.Cases.Cases
	runners := []Runner{}
	for _, cf := range cfs {
		if len(filters) > 0 {
			if !utils.ArrayContainsStringIgnoreCase(filters, cf.Name) {
				continue
			}
		}
		if strings.Contains(strings.ToLower(cf.Name), "influx") {
			if cof, err := cfg.FindInfluxDBConnection(cf.Connection); err == nil {
				if cli, err := client.NewInfluxClient(cof, ""); err == nil {
					r := NewInfluxRunner(cli, cf)
					runners = append(runners, &r)
				} else {
					logrus.WithError(err).Error("create runner failed")
				}
			}
		} else if strings.Contains(strings.ToLower(cf.Name), "mysql") {
			if cof, err := cfg.FindMySQLConnection(cf.Connection); err == nil {
				if cli, err := client.NewMySQLClient(cof); err == nil {
					if layout, err := mysql.GenerateLayout(pointsCfg.Measurement, pointsCfg.SeriesKey, pointsCfg.FieldsStr); err == nil {
						r := NewMySQLRunner(cli, cf, layout)
						runners = append(runners, &r)
					}
				} else {
					logrus.WithError(err).Error("create runner failed")
				}
			}
		}
	}
	return runners
}

func (r *caseRunner) doInsert(doWrite doWriteFunc) error {

	sink := stress.NewMultiSink(r.concurrency)
	sink.AddSink(stress.NewErrorSink(r.concurrency))

	if recordStats {
		sink.AddSink(stress.NewInfluxDBSink(int(r.concurrency), statsHost, statsDB))
	}

	sink.Open()

	var wg sync.WaitGroup
	wg.Add(r.concurrency)

	start := time.Now()

	totalWritten, err := doWrite(sink.Chan())
	r.totalWritten = totalWritten

	r.totalTime = time.Since(start)
	if err := r.cli.Close(); err != nil {
		logrus.WithError(err).Error("Error closing client")
	}

	sink.Close()
	r.throughput = r.totalWritten / uint64(r.totalTime.Seconds())
	if quiet {
		fmt.Println(r.throughput)
	} else {
		report.Append([]string{
			r.cfg.Name,
			r.cli.Connection(),
			"insert",
			fmt.Sprintf("%d", r.concurrency),
			fmt.Sprintf("%d", r.cfg.BatchSize),
			fmt.Sprintf("%s", start.Local().Format("2006-01-02 15:04:05")),
			fmt.Sprintf("%.3fs", r.totalTime.Seconds()),
			fmt.Sprintf("%d", r.throughput),
			fmt.Sprintf("%d", r.totalWritten)})
	}

	return err
}

func (r *caseRunner) Info() map[string]interface{} {
	return map[string]interface{}{
		"name":       r.cfg.Name,
		"connection": r.cli.Connection(),
	}
}

func (r *caseRunner) Result() map[string]interface{} {
	return map[string]interface{}{
		"throughput":    r.throughput,
		"total written": r.totalWritten,
		"total runtime": r.totalTime.Round(time.Second),
	}
}
