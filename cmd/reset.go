package cmd

import (
	"strings"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ()

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset database, drop all data",
	Long:  "",
	Run:   runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().StringVar(&dump, "dump", "", "Dump to given file instead of writing over HTTP")
}

func runReset(cmd *cobra.Command, args []string) {
	if strings.Contains(strings.ToLower(targets), "influx") {
		if err := resetInflux(config.Cfg); err == nil {
			logrus.Info("influxdb reseted")
		} else {
			logrus.WithError(err).Error("influxdb reset failed")
		}
	}
	if strings.Contains(strings.ToLower(targets), "mysql") {
		if err := resetMySQL(config.Cfg); err == nil {
			logrus.Info("mysql reseted")
		} else {
			logrus.WithError(err).Error("mysql reset failed")
		}
	}
}

func resetInflux(cfg config.Config) error {
	c, err := client.NewInfluxClient(dump)
	if err != nil {
		return err
	}
	return c.Reset()
}

func resetMySQL(cfg config.Config) error {
	c, err := client.NewMySQLClient(cfg.Connection.Mysql)
	if err != nil {
		return err
	}
	return c.Reset()
}
