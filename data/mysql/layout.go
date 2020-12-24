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
	return fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (id int auto_increment primary key, %s, create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(), INDEX time(create_time) %s);",
		l.name, l.genColumnDDL(), l.genIndexDDL())
}

func (l *Layout) genColumnDDL() string {
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
	for _, s := range l.tags {
		cols = append(cols, s[0]+" CHAR(32) NOT NULL DEFAULT ''")
	}
	if len(cols) > 0 {
		return strings.Join(cols, ", ")
	}
	return ""
}

func (l *Layout) genIndexDDL() string {
	ids := []string{}
	for _, s := range l.tags {
		ids = append(ids, s[0])
	}
	if len(ids) > 0 {
		return fmt.Sprintf(", INDEX idx_ss(%s)", strings.Join(ids, ", "))
	}
	return ""
}

func (l *Layout) genRow(str string) Row {
	r := Row{}
	for range l.ints {
		r.AppendCol(utils.RandInt31Safe())
	}
	for range l.floats {
		r.AppendCol(utils.RandInt31Safe())
	}
	for range l.strs {
		r.AppendCol("'" + str + "'")
	}
	for _, t := range l.tags {
		r.AppendCol(fmt.Sprintf("'%s-%d'", t[1], utils.RandInt31nSafe(300)))
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
