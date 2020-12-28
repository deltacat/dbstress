package cmd

import (
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
}

func runReset(cmd *cobra.Command, args []string) {
	// reset all influx
	for _, cc := range config.Cfg.Connection.InfluxDB {
		logger := logrus.WithField("connection", cc.Name)
		if err := resetInflux(cc); err == nil {
			logger.Info("influxdb reseted")
		} else {
			logger.WithError(err).Error("influxdb reset failed")
		}

	}
	// reset all mysql
	for _, cc := range config.Cfg.Connection.MySQL {
		logger := logrus.WithField("connection", cc.Name)
		if err := resetMySQL(cc); err == nil {
			logger.Info("mysql reseted")
		} else {
			logger.WithError(err).Error("mysql reset failed")
		}
	}
}

func resetInflux(cc config.InfluxClientConfig) error {
	c, err := client.NewInfluxClient(cc, dump)
	if err != nil {
		return err
	}
	return c.Reset()
}

func resetMySQL(cc config.MySQLClientConfig) error {
	c, err := client.NewMySQLClient(cc)
	if err != nil {
		return err
	}
	return c.Reset()
}
