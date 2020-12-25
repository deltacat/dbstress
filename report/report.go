package report

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

var table *tablewriter.Table

// SetHeader set report table header row
func SetHeader(keys []string) {
	table.SetHeader(keys)
}

// Append append row to report
func Append(row []string) {
	table.Append(row)
}

// Render render report
func Render() {
	table.Render()
}

func init() {
	table = tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
}
