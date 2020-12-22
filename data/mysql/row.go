package mysql

import (
	"fmt"
	"strings"
	"time"
)

// Row mysql table row
type Row struct {
	time    time.Time
	colVals []string
}

// SetTime ...
func (r *Row) SetTime(t time.Time) {

}

// Update ...
func (r *Row) Update() {

}

func (r *Row) getInsertSegments() string {
	return fmt.Sprintf("(%s)", strings.Join(r.colVals, ","))
}

// AppendCol append column data
func (r *Row) AppendCol(c interface{}) {
	r.colVals = append(r.colVals, fmt.Sprintf("%v", c))
}
