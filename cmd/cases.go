package cmd

import (
	"strings"
	"time"

	"github.com/deltacat/dbstress/runner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var caseCmd = &cobra.Command{
	Use:   "cases",
	Short: "run predefined cases",
	Long:  "",
	Run:   runCases,
}

var (
	casesToRunStr string
)

func init() {
	rootCmd.AddCommand(caseCmd)

	caseCmd.Flags().StringVarP(&casesToRunStr, "run", "r", "", "Select cases to run. Default all cases")
}

func runCases(cmd *cobra.Command, args []string) {
	runner.Setup(tick, fast, quiet, kapacitorMode, cfg.Points, cfg.StatsRecord)
	defer runner.Close()

	casesToRun := []string{}
	if casesToRunStr != "" {
		casesToRun = strings.Split(casesToRunStr, ",")
	}

	runners := runner.BuildAllRunners(cfg, casesToRun)
	logrus.WithField("cases", casesToRun).WithField("build", len(runners)).Infof("build runner from cases config, start run")

	delay := cfg.Cases.Delay
	for i, r := range runners {
		if i > 0 {
			logrus.WithField("delay", delay).Info("finished case, delay for next")
			<-time.Tick(delay)
		}
		logrus.WithFields(logrus.Fields(r.Info())).Info("running case")
		r.Run()
	}
}
