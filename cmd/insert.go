package cmd

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/data/influx/lineprotocol"
	"github.com/deltacat/dbstress/data/influx/point"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/deltacat/dbstress/report"
	"github.com/deltacat/dbstress/stress"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	layout                          mysql.Layout
	dump                            string
	seriesN                         int
	concurrency, batchSize, pointsN uint64
	runtime                         time.Duration
)

var insertCmd = &cobra.Command{
	Use:   "insert SERIES FIELDS",
	Short: "Insert data into DB",
	Long:  "",
	Run:   runInsert,
}

func init() {
	rootCmd.AddCommand(insertCmd)

	insertCmd.Flags().IntVarP(&seriesN, "series", "s", 100000, "number of series that will be written")
	insertCmd.Flags().Uint64VarP(&pointsN, "points", "n", math.MaxUint64, "number of points that will be written")
	insertCmd.Flags().Uint64VarP(&batchSize, "batch-size", "b", 10000, "number of points in a batch")
	insertCmd.Flags().DurationVarP(&runtime, "runtime", "r", time.Duration(math.MaxInt64), "Total time that the test will run")
	insertCmd.Flags().StringVar(&dump, "dump", "", "Dump to given file instead of writing over HTTP")
}

func runInsert(cmd *cobra.Command, args []string) {

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

	report.SetHeader([]string{"client", "action", "run", "Throughput", "Points"})

	logrus.Info("will insert to influxdb")
	if err := insertInflux(); err != nil {
		logrus.WithError(err).Error("error with inserting influxdb")
	}
	logrus.Info("will insert to mysql")
	if err := insertMysql(); err != nil {
		logrus.WithError(err).Error("error with inserting mysql")
	}

	if !quiet {
		fmt.Printf("\nReport: =======>\n")
		fmt.Printf("Use point template: %s %s <timestamp>\n", seriesKey, fieldStr)
		fmt.Printf("Use batch size of %d line(s)\n", batchSize)
		fmt.Printf("Spreading writes across %d series\n", seriesN)
		fmt.Printf("Use %d concurrent writer(s)\n", concurrency)
		report.Render()
	}

}

func insertMysql() error {
	cc, err := cfg.FindDefaultMySQLConnection()
	if err != nil {
		return err
	}

	cli, err := client.NewMySQLClient(cc)
	defer cli.Close()
	if err != nil {
		return err
	}

	layout, err = mysql.GenerateLayout(measurement, seriesKey, fieldStr)
	if err != nil {
		return err
	}

	if err = cli.Create(layout.GetCreateStmt()); err != nil {
		return err
	}

	return doInsert(cli, doWriteMysql)
}

func insertInflux() error {

	c, err := cfg.FindDefaultInfluxDBConnection()
	if err != nil {
		return err
	}

	cli, _ := client.NewInfluxClient(c, dump)
	defer cli.Close()
	if !kapacitorMode {
		if err := cli.Create(""); err != nil {
			return err
		}
	}

	return doInsert(cli, doWriteInflux)
}

type doWriteFunc func(cli client.Client, resultChan chan stress.WriteResult) (uint64, error)

func doInsert(cli client.Client, doWrite doWriteFunc) error {

	sink := stress.NewMultiSink(int(concurrency))
	sink.AddSink(stress.NewErrorSink(int(concurrency)))

	if recordStats {
		sink.AddSink(stress.NewInfluxDBSink(int(concurrency), statsHost, statsDB))
	}

	sink.Open()

	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	start := time.Now()

	totalWritten, err := doWrite(cli, sink.Chan())

	totalTime := time.Since(start)
	if err := cli.Close(); err != nil {
		logrus.WithError(err).Error("Error closing client")
	}

	sink.Close()
	throughput := int(float64(totalWritten) / totalTime.Seconds())
	if quiet {
		fmt.Println(throughput)
	} else {

		report.Append([]string{
			cli.Name(),
			"insert",
			fmt.Sprintf("%.3fs", totalTime.Seconds()),
			fmt.Sprintf("%d", throughput),
			fmt.Sprintf("%d", totalWritten)})

		logrus.WithField("Write Throughput:", throughput).WithField("Points Written:", totalWritten).Info("run stress done")
	}

	return err
}

func doWriteInflux(cli client.Client, resultChan chan stress.WriteResult) (uint64, error) {
	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	var totalWritten uint64
	startSplit := 0
	inc := int(seriesN) / int(concurrency)
	endSplit := inc

	pts := point.NewPoints(measurement, seriesKey, fieldStr, seriesN, lineprotocol.Nanosecond)
	for i := uint64(0); i < concurrency; i++ {

		go func(startSplit, endSplit int) {
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: batchSize,
				MaxPoints: pointsN / concurrency, // divide by concurreny
				GzipLevel: cli.GzipLevel(),
				Deadline:  time.Now().Add(runtime),
				Tick:      tick,
				Results:   resultChan,
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.WriteInflux(pts[startSplit:endSplit], cli, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()

	return totalWritten, nil
}

func doWriteMysql(cli client.Client, resultChan chan stress.WriteResult) (uint64, error) {

	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	totalWritten := uint64(0)
	startSplit := 0
	inc := int(seriesN) / int(concurrency)
	endSplit := inc

	tbl := mysql.NewTableChunk(layout, batchSize)
	for i := uint64(0); i < concurrency; i++ {

		go func(startSplit, endSplit int) {
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: batchSize,
				MaxPoints: pointsN / concurrency, // divide by concurreny
				Deadline:  time.Now().Add(runtime),
				Tick:      tick,
				Results:   resultChan,
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.WriteMySQL(tbl, cli, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()

	return totalWritten, nil
}
