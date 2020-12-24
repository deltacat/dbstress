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
	resetCmd.Flags().StringVar(&dump, "dump", "", "Dump to given file instead of writing over HTTP")
}

func runReset(cmd *cobra.Command, args []string) {
	if err := resetInflux(config.Cfg); err == nil {
		logrus.Info("influxdb reseted")
	} else {
		logrus.WithError(err).Error("influxdb reset failed")
	}
	if err := resetMySQL(config.Cfg); err == nil {
		logrus.Info("mysql reseted")
	} else {
		logrus.WithError(err).Error("mysql reset failed")
	}
}

func resetInflux(cfg config.Config) error {
	cc, err := cfg.FindDefaultInfluxDBConnection()
	if err != nil {
		return err
	}

	c, err := client.NewInfluxClient(cc, dump)
	if err != nil {
		return err
	}
	return c.Reset()
}

func resetMySQL(cfg config.Config) error {
	cc, err := cfg.FindDefaultMySQLConnection()
	if err != nil {
		return err
	}
	c, err := client.NewMySQLClient(cc)
	if err != nil {
		return err
	}
	return c.Reset()
}
