package point

import (
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/lineprotocol"
)

// The point struct implements the lineprotocol.Point interface.
type point struct {
	seriesKey []byte

	// Note here that Ints and Floats are exported so they can be modified outside
	// of the point struct
	Ints    []*lineprotocol.Int
	Floats  []*lineprotocol.Float
	Strings []*lineprotocol.String

	// The fields slice should contain exactly Ints and Floats. Having this
	// slice allows us to avoid iterating through Ints and Floats in the Fields
	// function.
	fields []lineprotocol.Field

	time *lineprotocol.Timestamp
}

// returns a new point without setting the time field.
func build(sk []byte, ints, floats []string, strs []string, p lineprotocol.Precision) *point {
	fields := []lineprotocol.Field{}
	e := &point{
		seriesKey: sk,
		time:      lineprotocol.NewTimestamp(p),
		fields:    fields,
	}

	for _, i := range ints {
		n := &lineprotocol.Int{Key: []byte(i)}
		e.Ints = append(e.Ints, n)
		e.fields = append(e.fields, n)
	}

	for _, f := range floats {
		n := &lineprotocol.Float{Key: []byte(f)}
		e.Floats = append(e.Floats, n)
		e.fields = append(e.fields, n)
	}

	for _, s := range strs {
		n := &lineprotocol.String{Key: []byte(s), Value: genString()}
		e.Strings = append(e.Strings, n)
		e.fields = append(e.fields, n)
	}

	return e
}

// rand string is time expensive
func genString() string {
	// len := 64
	// return randstr.String(len)
	return "zze6TQ2TfpJPb0UVLs3FckJtuXhTQVVNIFtTrJEEWoFJxFukX3alzbiV2dq4RidR"
}

// Series returns the series key for a point.
func (p *point) Series() []byte {
	return p.seriesKey
}

// Fields returns the fields for a a point.
func (p *point) Fields() []lineprotocol.Field {
	return p.fields
}

// Time returns the timestamps for a point.
func (p *point) Time() *lineprotocol.Timestamp {
	return p.time
}

// SetTime set the t to be the timestamp for a point.
func (p *point) SetTime(t time.Time) {
	p.time.SetTime(&t)
}

// Update increments the value of all of the Int and Float
// fields by 1.
func (p *point) Update() {
	for _, i := range p.Ints {
		atomic.AddInt64(&i.Value, int64(1))
	}

	for _, f := range p.Floats {
		// Need to do something else here
		// There will be a race here
		f.Value += 1.0
	}

	newStr := genString()
	for _, s := range p.Strings {
		s.Value = newStr
	}
}

// NewPoints returns a slice of Points of length seriesN shaped like the given seriesKey.
func NewPoints(measurement, seriesKey, fields string, seriesN int, pc lineprotocol.Precision) []lineprotocol.Point {
	pts := []lineprotocol.Point{}
	series := generateSeriesKeys(measurement, seriesKey, seriesN)
	ints, floats, strs := generateFieldSet(fields)
	for _, sk := range series {
		p := build(sk, ints, floats, strs, pc)
		pts = append(pts, p)
	}

	return pts
}
