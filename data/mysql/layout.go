package mysql

import (
	"fmt"
	"strings"

	"github.com/deltacat/dbstress/data/fieldset"
	"github.com/deltacat/dbstress/utils"
)

// Layout mysql table layout definition
type Layout struct {
	name    string
	allCols []string
	ints    []string
	floats  []string
	strs    []string
	tags    [][]string // 对照 influxdb 的 tags 生成 mysql 索引，元素[0]为column名，[1]为值前缀
}

// GenInsertStmtValues generate insert row DML
func (l *Layout) GenInsertStmtValues(colVals []string) string {
	return fmt.Sprintf("(null,%s,now())", strings.Join(colVals, ","))
}

// GetCreateStmt get create table DDL
func (l *Layout) GetCreateStmt() string {
	fields := l.genNormalColumnDDL()
	tags := l.genIndexColumnDDL()
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (id int auto_increment primary key, 
		%s, 
		create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
		INDEX time (create_time)
		);`,
		l.name, strings.Join(append(fields, tags...), ","))
}

func (l *Layout) genNormalColumnDDL() []string {
	cols := []string{}
	for _, s := range l.ints {
		cols = append(cols, s+" INT")
	}
	for _, s := range l.floats {
		cols = append(cols, s+" FLOAT")
	}
	for _, s := range l.strs {
		cols = append(cols, s+" CHAR(64)")
	}
	return cols
}

func (l *Layout) genIndexColumnDDL() []string {
	cols := []string{}
	for _, s := range l.tags {
		cols = append(cols, s[0]+" CHAR(32)")
	}
	return cols
}

func (l *Layout) genRow(str string) Row {
	r := Row{}
	for range l.ints {
		r.AppendCol(utils.RandIntSafe())
	}
	for range l.floats {
		r.AppendCol(utils.RandIntSafe())
	}
	for range l.strs {
		r.AppendCol("'" + str + "'")
	}
	for _, t := range l.tags {
		r.AppendCol("'" + t[1] + "'")
	}
	return r
}

// GenerateLayout generate a new layout
func GenerateLayout(measurement, tagsStr, fieldsStr string) (Layout, error) {
	ints, floats, strs := fieldset.GenerateFieldSet(fieldsStr)
	tags := fieldset.GenerateTagsSet(tagsStr)
	return Layout{
		name:   measurement,
		ints:   ints,
		floats: floats,
		strs:   strs,
		tags:   tags,
	}, nil
}
