package cmd

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCasesCmd = &cobra.Command{
	Use:   "list",
	Short: "list all cases",
	Long:  "",
	Run:   listCases,
}

func init() {
	caseCmd.AddCommand(listCasesCmd)
}

func listCases(cmd *cobra.Command, args []string) {
	tw := tablewriter.NewWriter(os.Stdout)
	for _, cc := range cfg.Cases.Cases {
		tw.Append([]string{
			cc.Name,
			cc.Connection,
			strconv.FormatInt(int64(cc.Concurrent), 10),
			strconv.FormatInt(int64(cc.BatchSize), 10),
			strconv.FormatInt(int64(cc.Runtime.Seconds()), 10),
		})

	}
	if tw.NumLines() > 0 {
		tw.SetHeader([]string{"name", "connection", "concurrency", "batch", "run"})
		tw.Render()
	}
}
