package cmd

import (
	"time"

	"github.com/deltacat/dbstress/runner"
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
	defer runner.Close()

	runners := runner.BuildAllRunners(cfg)
	logrus.WithField("found", len(cfg.Cases.Cases)).WithField("build", len(runners)).Infof("build runner from cases config, start run")

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
