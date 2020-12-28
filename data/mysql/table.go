package mysql

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/deltacat/dbstress/utils"
)

// TableChunk table data chunk struct
type TableChunk struct {
	layout     Layout
	rows       []Row
	strValue   string
	intValue   int
	floatValue float32
	rr         *rand.Rand
}

// GetRowsNum get number of rows
func (t *TableChunk) GetRowsNum() uint64 {
	return uint64(len(t.rows))
}

// GenInsertStmt get statement of insertion all rows
func (t *TableChunk) GenInsertStmt() string {

	segs := []string{}
	for _, v := range t.rows {
		segs = append(segs, t.layout.GenInsertStmtValues(v.GetColVals()))
	}

	stmt := fmt.Sprintf("INSERT INTO %s VALUES %s;", t.layout.name, strings.Join(segs, ","))
	return stmt
}

// Update update table data
func (t *TableChunk) Update() {
	t.strValue = utils.RandStrSafe(utils.StrDataLength)
	t.intValue++
	t.floatValue += 0.1
	for i := range t.rows {
		t.rows[i] = t.layout.genRow(t.intValue, t.floatValue, t.strValue, t.rr)
	}
}

// NewTableChunk generate batch rows
func NewTableChunk(layout Layout, batchSize uint64) TableChunk {
	rows := make([]Row, int(batchSize))
	t := TableChunk{
		layout: layout,
		rows:   rows,
		rr:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return t
}
