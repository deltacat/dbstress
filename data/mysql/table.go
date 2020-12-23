package mysql

import (
	"fmt"
	"strings"

	"github.com/deltacat/dbstress/utils"
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
		segs = append(segs, t.layout.GenInsertStmtValues(v.GetColVals()))
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
	str := utils.RandStrSafe(64)
	for i := range t.rows {
		t.rows[i] = t.layout.genRow(str)
	}
}

// NewTableChunk generate batch rows
func NewTableChunk(layout Layout, batchSize uint64) Table {
	rows := make([]Row, int(batchSize))
	t := Table{
		layout: layout,
		rows:   rows,
	}
	t.Update()
	return t
}
