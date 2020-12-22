package mysql

import (
	"fmt"
	"strings"
)

// Table table struct
type Table struct {
	layout Layout
	rows   []Row
}

// GetRowsNum get number of rows
func (t *Table) GetRowsNum() uint64 {
	return uint64(len(t.rows))
}

// GenInsertStmt get statement of insertion all rows
func (t *Table) GenInsertStmt() string {

	segs := []string{}
	for _, v := range t.rows {
		segs = append(segs, v.getInsertSegments())
	}

	stmt := fmt.Sprintf("INSERT INTO %s VALUES %s;", t.layout.name, strings.Join(segs, ","))
	return stmt
}

func (t *Table) buildRow() Row {
	r := Row{}
	for _, c := range t.layout.ints {
		r.AppendCol(c)
	}
	return r
}

// Update update table data
func (t *Table) Update() {

}

// GenBatchTableData generate batch rows
func GenBatchTableData(layout Layout, batchSize uint64) Table {

	rows := make([]Row, int(batchSize))
	for i := range rows {
		rows[i] = layout.genRow()
	}

	return Table{
		layout: layout,
		rows:   rows,
	}
}
