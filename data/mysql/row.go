package mysql

import (
	"fmt"
)

// Row mysql table row
type Row struct {
	colVals []string
}

// GetColVals return column values via string
func (r *Row) GetColVals() []string {
	return r.colVals
}

// AppendCol append column data
func (r *Row) AppendCol(c interface{}) {
	r.colVals = append(r.colVals, fmt.Sprintf("%v", c))
}
