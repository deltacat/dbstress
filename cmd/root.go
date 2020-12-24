package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/deltacat/dbstress/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "dbstress",
	Short:            "Create artificial load on an InfluxDB/MySQL instance",
	Long:             "This application create stress test on influxdb or mysql.\nPlease rename dbstress.sample.toml to dbstress.toml then make necessary change",
	PersistentPreRun: runRootPersistentPre,
}

var (
	cfg                              config.Config // global configure holder
	statsHost, statsDB               string
	recordStats                      bool
	measurement, seriesKey, fieldStr string
	pps                              uint64
	tick                             time.Duration
	fast, quiet                      bool
	strict, kapacitorMode            bool
	tlsSkipVerify                    bool
)

// Execute run root cmd
func Execute(v VersionInfo) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	setDefaultConfig()

	rootCmd.PersistentFlags().StringVarP(&statsHost, "stats-host", "", "http://localhost:8086", "Address of InfluxDB instance where runtime statistics will be recorded")
	rootCmd.PersistentFlags().StringVarP(&statsDB, "stats-db", "", "stress_stats", "Database that statistics will be written to")
	rootCmd.PersistentFlags().BoolVarP(&recordStats, "stats", "", false, "Record runtime statistics")

	rootCmd.PersistentFlags().Uint64VarP(&pps, "pps", "", 200000, "Points Per Second")
	rootCmd.PersistentFlags().DurationVarP(&tick, "tick", "", time.Second, "Amount of time between request")
	rootCmd.PersistentFlags().BoolVarP(&fast, "fast", "f", false, "Run as fast as possible")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Only print the write throughput")
	rootCmd.PersistentFlags().BoolVarP(&kapacitorMode, "kapacitor", "k", false, "Use Kapacitor mode, namely do not try to run any queries.")
	rootCmd.PersistentFlags().BoolVarP(&strict, "strict", "", false, "Strict mode will exit as soon as an error or unexpected status is encountered")

}

func runRootPersistentPre(cmd *cobra.Command, args []string) {
	cfg = config.Cfg
	measurement = cfg.Points.Measurement
	seriesKey = cfg.Points.SeriesKey
	fieldStr = cfg.Points.FieldsStr

	if !strings.Contains(seriesKey, ",") && !strings.Contains(seriesKey, "=") {
		logrus.Warnf("expect series like 'ctr,some=tag', got '%s'", seriesKey)
		os.Exit(1)
		return
	}
}
