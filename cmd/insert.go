package cmd

import (
	"fmt"
	"math"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/deltacat/dbstress/runner"
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

	runner.Setup(tick, fast, quiet, kapacitorMode, cfg.Points, cfg.StatsRecord)
	defer runner.Close()

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

	logrus.Info("will insert to influxdb")
	if err := insertInflux(); err != nil {
		logrus.WithError(err).Error("error with inserting influxdb")
	}
	logrus.Info("will insert to mysql")
	if err := insertMysql(); err != nil {
		logrus.WithError(err).Error("error with inserting mysql")
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

	cs := config.CaseConfig{
		Name:       "Insert MySQL",
		Connection: cc.Name,
		Concurrent: int(concurrency),
		BatchSize:  int(batchSize),
		Runtime:    runtime,
	}

	r := runner.NewMySQLRunner(cli, cs, layout)
	return r.Run()
}

func insertInflux() error {
	cc, err := cfg.FindDefaultInfluxDBConnection()
	if err != nil {
		return err
	}

	cli, _ := client.NewInfluxClient(cc, dump)
	defer cli.Close()
	if !kapacitorMode {
		if err := cli.Create(""); err != nil {
			return err
		}
	}

	cs := config.CaseConfig{
		Name:       "Insert Influx",
		Connection: cc.Name,
		Concurrent: int(concurrency),
		BatchSize:  int(batchSize),
		Runtime:    runtime,
	}
	r := runner.NewInfluxRunner(cli, cs)

	return r.Run()
}
