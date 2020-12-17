package cmd

import (
	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Test mysql connection",
	Long:  "Test mysql connection. Just for test, will removed later",
	Run:   runMysqlCmd,
}

func init() {
	rootCmd.AddCommand(mysqlCmd)
}

func runMysqlCmd(cmd *cobra.Command, args []string) {
	logrus.Info("run mysql cmd")
	if cli, err := client.NewMySQLClient(config.Cfg.Connection.Mysql); err != nil {
		logrus.WithError(err).Error("create mysql client failed")
	} else {
		cli.Close()
	}
}
