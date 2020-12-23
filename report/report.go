package report

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

var table *tablewriter.Table = tablewriter.NewWriter(os.Stdout)

func SetHeader(keys []string) {
	table.SetHeader(keys)

}

func Append(row []string) {
	table.Append(row)
}

func Render() {
	table.Render()
}
