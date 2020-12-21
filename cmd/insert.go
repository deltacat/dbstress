package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/data/influx/lineprotocol"
	"github.com/deltacat/dbstress/data/influx/point"
	"github.com/deltacat/dbstress/stress"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	measurement, seriesKey, fieldStr     string
	statsHost, statsDB                   string
	dump                                 string
	seriesN                              int
	concurrency, batchSize, pointsN, pps uint64
	runtime                              time.Duration
	tick                                 time.Duration
	fast, quiet                          bool
	strict, kapacitorMode                bool
	recordStats                          bool
	tlsSkipVerify                        bool
)

var insertCmd = &cobra.Command{
	Use:   "insert SERIES FIELDS",
	Short: "Insert data into DB",
	Long:  "",
	Run:   runInsert,
}

func init() {
	rootCmd.AddCommand(insertCmd)

	insertCmd.Flags().StringVarP(&statsHost, "stats-host", "", "http://localhost:8086", "Address of InfluxDB instance where runtime statistics will be recorded")
	insertCmd.Flags().StringVarP(&statsDB, "stats-db", "", "stress_stats", "Database that statistics will be written to")
	insertCmd.Flags().BoolVarP(&recordStats, "stats", "", false, "Record runtime statistics")

	insertCmd.Flags().IntVarP(&seriesN, "series", "s", 100000, "number of series that will be written")
	insertCmd.Flags().Uint64VarP(&pointsN, "points", "n", math.MaxUint64, "number of points that will be written")
	insertCmd.Flags().Uint64VarP(&batchSize, "batch-size", "b", 10000, "number of points in a batch")
	insertCmd.Flags().Uint64VarP(&pps, "pps", "", 200000, "Points Per Second")
	insertCmd.Flags().DurationVarP(&runtime, "runtime", "r", time.Duration(math.MaxInt64), "Total time that the test will run")
	insertCmd.Flags().DurationVarP(&tick, "tick", "", time.Second, "Amount of time between request")
	insertCmd.Flags().BoolVarP(&fast, "fast", "f", false, "Run as fast as possible")
	insertCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only print the write throughput")
	insertCmd.Flags().BoolVarP(&kapacitorMode, "kapacitor", "k", false, "Use Kapacitor mode, namely do not try to run any queries.")
	insertCmd.Flags().StringVar(&dump, "dump", "", "Dump to given file instead of writing over HTTP")
	insertCmd.Flags().BoolVarP(&strict, "strict", "", false, "Strict mode will exit as soon as an error or unexpected status is encountered")
}

func runInsert(cmd *cobra.Command, args []string) {

	cfg := config.Cfg
	measurement = cfg.Points.Measurement
	seriesKey = cfg.Points.SeriesKey
	fieldStr = cfg.Points.FieldsStr

	if !strings.Contains(seriesKey, ",") && !strings.Contains(seriesKey, "=") {
		logrus.Warnf("expect series like 'ctr,some=tag', got '%s'", seriesKey)
		os.Exit(1)
		return
	}

	concurrency = pps / batchSize
	// PPS takes precedence over batchSize.
	// Adjust accordingly.
	if pps < batchSize {
		batchSize = pps
		concurrency = 1
	}
	if !quiet {
		fmt.Printf("Using point template: %s %s <timestamp>\n", seriesKey, fieldStr)
		fmt.Printf("Using batch size of %d line(s)\n", batchSize)
		fmt.Printf("Spreading writes across %d series\n", seriesN)
		if fast {
			fmt.Println("Output is unthrottled")
		} else {
			fmt.Printf("Throttling output to ~%d points/sec\n", pps)
		}
		fmt.Printf("Using %d concurrent writer(s)\n", concurrency)

		fmt.Printf("Running until ~%d points sent or until ~%v has elapsed\n", pointsN, runtime)
	}

	if strings.Contains(strings.ToLower(targets), "influx") {
		logrus.Info("will insert to influxdb")
		insertInflux(cfg)
	}

	if strings.Contains(strings.ToLower(targets), "mysql") {
		logrus.Info("will insert to mysql")
		insertMysql(cfg)
	}

}

func insertMysql(cfg config.Config) {
	if cli, err := client.NewMySQLClient(cfg.Connection.Mysql); err != nil {
		logrus.WithError(err).Error("create mysql client failed")
	} else {
		if err := cli.Create(); err != nil {
			logrus.WithError(err).Error("create database failed")
		}
		cli.Close()
	}
}

func insertInflux(cfg config.Config) {

	c := client.NewInfluxClient(dump)

	if !kapacitorMode {
		if err := c.Create(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create database:", err.Error())
			fmt.Fprintln(os.Stderr, "Aborting.")
			os.Exit(1)
			return
		}
	}

	pts := point.NewPoints(measurement, seriesKey, fieldStr, seriesN, lineprotocol.Nanosecond)

	startSplit := 0
	inc := int(seriesN) / int(concurrency)
	endSplit := inc

	sink := stress.NewMultiSink(int(concurrency))
	sink.AddSink(stress.NewErrorSink(int(concurrency)))

	if recordStats {
		sink.AddSink(stress.NewInfluxDBSink(int(concurrency), statsHost, statsDB))
	}

	sink.Open()

	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	var totalWritten uint64

	start := time.Now()
	gzip := cfg.Connection.Influxdb.Gzip
	for i := uint64(0); i < concurrency; i++ {

		go func(startSplit, endSplit int) {
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: batchSize,
				MaxPoints: pointsN / concurrency, // divide by concurreny
				GzipLevel: gzip,
				Deadline:  time.Now().Add(runtime),
				Tick:      tick,
				Results:   sink.Chan(),
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.WriteInflux(pts[startSplit:endSplit], c, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()
	totalTime := time.Since(start)
	if err := c.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing client: %v\n", err.Error())
	}

	sink.Close()
	throughput := int(float64(totalWritten) / totalTime.Seconds())
	if quiet {
		fmt.Println(throughput)
	} else {
		fmt.Println("Write Throughput:", throughput)
		fmt.Println("Points Written:", totalWritten)
	}
}
