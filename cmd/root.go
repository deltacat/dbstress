package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dbstress",
	Short: "Create artificial load on an InfluxDB/MySQL instance",
	Long:  "This application create stress test on influxdb or mysql.\nPlease rename dbstress.sample.toml to dbstress.toml then make necessary change",
}

var targets string

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

	rootCmd.PersistentFlags().StringVarP(&targets, "targets", "t", "influx", "set which target to run (mysql, influxdb)")
}
