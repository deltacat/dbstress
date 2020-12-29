package cmd

import (
	"errors"
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
	runner.Setup(cfg.Cases.Tick, cfg.Cases.Fast, quiet, kapacitorMode, cfg.Points, cfg.StatsRecord)
	defer runner.Close()

	casesToRun := []string{}
	if casesToRunStr != "" {
		casesToRun = strings.Split(casesToRunStr, ",")
	}

	runners := runner.BuildAllRunners(cfg, casesToRun)
	if len(runners) == 0 {
		logrus.Warnln("no valid case to run")
		return
	}
	defer runner.Report()

	logrus.WithField("cases", casesToRun).WithField("build", len(runners)).Infof("build runner from cases config, start run")

	delay := cfg.Cases.Delay
	for i, r := range runners {
		logrus.WithFields(logrus.Fields(r.Info())).Infof("running case %d/%d", i+1, len(runners))
		err := r.Run()
		logger := logrus.WithFields(logrus.Fields(r.Result()))
		if err != nil {
			logger = logrus.WithError(errors.Unwrap(err))
		}
		if i == len(runners)-1 {
			logger.Info("finished case")
		} else {
			logger.WithField("wait", delay).Info("finished case, wait a while before next")
			<-time.Tick(delay)
		}
	}
}
