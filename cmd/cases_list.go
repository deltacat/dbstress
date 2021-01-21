package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/deltacat/dbstress/csv"
	"github.com/deltacat/dbstress/runner"
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
	caseCmd.AddCommand(genCasesCmd)
}

func listCases(cmd *cobra.Command, args []string) {
	tw := tablewriter.NewWriter(os.Stdout)
	for _, cc := range loadCases() {
		tw.Append([]string{
			cc.Name,
			cc.Connection,
			strconv.FormatInt(int64(cc.Concurrent), 10),
			strconv.FormatInt(int64(cc.BatchSize), 10),
			cc.Runtime.String(),
		})

	}
	if tw.NumLines() > 0 {
		tw.SetHeader([]string{"name", "connection", "concur", "batch", "run"})
		tw.Render()
	}
}

var genCasesCmd = &cobra.Command{
	Use:   "gen",
	Short: "gen cases file sample",
	Long:  "",
	Run:   genCases,
}

func genCases(cmd *cobra.Command, args []string) {
	tmpl := []runner.CaseConfig{{
		Name:       "Sample",
		Connection: "Influx1.8",
		Concurrent: 20,
		BatchSize:  2000,
		Runtime:    csv.Duration{Duration: time.Second * 30},
	}}

	csv.Output("./cases.csv", tmpl)
}
