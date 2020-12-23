package cmd

import (
	"fmt"

	"github.com/deltacat/dbstress/config"
	"github.com/spf13/cobra"
)

// VersionInfo version info, would be injected while making
type VersionInfo struct {
	Project, Version, Timestamp, Revision string
}

var version VersionInfo

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the app version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n%s, %s\nversion:  %s\nrevision: %s\ntime:     %s\n", version.Project, rootCmd.Short, version.Version, version.Revision, version.Timestamp)
		fmt.Printf("config:   %+v\n", config.Cfg)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
