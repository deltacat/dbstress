package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/cmd/runner"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/deltacat/dbstress/report"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var caseCmd = &cobra.Command{
	Use:   "cases [case[,case...]]",
	Short: "run predefined cases",
	Long:  "",
	Run:   runCases,
}

func init() {
	rootCmd.AddCommand(caseCmd)
}

func runCases(cmd *cobra.Command, args []string) {

	runner.Setup(tick, fast, quiet, kapacitorMode, cfg.Points)

	delay := cfg.Cases.Delay
	cfs := cfg.Cases.Cases

	runners := []runner.Runner{}
	for _, cf := range cfs {
		if strings.Contains(strings.ToLower(cf.Name), "influx") {
			if cof, err := cfg.FindInfluxDBConnection(cf.Connection); err == nil {
				if cli, err := client.NewInfluxClient(cof, ""); err == nil {
					r := runner.NewInfluxRunner(cli, cf)
					runners = append(runners, &r)
				}
			}
		} else if strings.Contains(strings.ToLower(cf.Name), "mysql") {
			if cof, err := cfg.FindMySQLConnection(cf.Connection); err == nil {
				if cli, err := client.NewMySQLClient(cof); err == nil {
					if layout, err = mysql.GenerateLayout(measurement, seriesKey, fieldStr); err == nil {
						r := runner.NewMySQLRunner(cli, cf, layout)
						runners = append(runners, &r)
					}
				}
			}
		}
	}

	logrus.WithField("found", len(cfs)).WithField("build", len(runners)).Infof("build runner from cases config, start run")

	for i, r := range runners {
		if i > 0 {
			logrus.WithField("delay", delay).Info("delay for next case")
			<-time.Tick(delay)
		}
		r.Run()
	}

	if !quiet {
		fmt.Printf("\nReport: =======>\n")
		fmt.Printf("Use point template: %s %s <timestamp>\n", seriesKey, fieldStr)
		fmt.Printf("Use batch size of %d line(s)\n", batchSize)
		fmt.Printf("Spreading writes across %d series\n", seriesN)
		fmt.Printf("Use %d concurrent writer(s)\n", concurrency)
		report.SetHeader([]string{"case", "connection", "action", "concurrency", "batch size", "run", "throughput", "points"})
		report.Render()
	}

}
