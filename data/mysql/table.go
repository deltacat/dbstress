package mysql

import (
	"fmt"
	"strings"

	"github.com/deltacat/dbstress/utils"
)

// Table table struct
type Table struct {
	layout     Layout
	rows       []Row
	strValue   string
	intValue   int
	floatValue float32
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

// Update update table data
func (t *Table) Update() {
	t.strValue = utils.RandStrSafe(utils.StrDataLength)
	t.intValue++
	t.floatValue += 0.1
	for i := range t.rows {
		t.rows[i] = t.layout.genRow(t.intValue, t.floatValue, t.strValue)
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
