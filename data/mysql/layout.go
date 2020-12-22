package mysql

import (
	"fmt"
	"strings"

	"github.com/deltacat/dbstress/data/fieldset"
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

// GenInsertStmt generate insert row DML
func (l *Layout) GenInsertStmt(rows []Row) string {
	return ""
}

// GetCreateStmt get create table DDL
func (l *Layout) GetCreateStmt() string {
	fields := l.genNormalColumnDDL()
	tags := l.genIndexColumnDDL()
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s)",
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
		cols = append(cols, s+" VARCHAR(64)")
	}
	return cols
}

func (l *Layout) genIndexColumnDDL() []string {
	cols := []string{}
	for _, s := range l.tags {
		cols = append(cols, s[0]+" VARCHAR(64)")
	}
	return cols
}

func (l *Layout) genRow() Row {
	r := Row{}
	for range l.ints {
		r.AppendCol(0)
	}
	for range l.floats {
		r.AppendCol(0.1)
	}
	for range l.strs {
		r.AppendCol("'zze6TQ2TfpJPb0UVLs3FckJtuXhTQVVNIFtTrJEEWoFJxFukX3alzbiV2dq4RidR'")
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
