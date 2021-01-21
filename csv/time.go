package csv

import "time"

// Duration for customize marshal
type Duration struct {
	time.Duration
}

// UnmarshalCSV convert the CSV string as internal duration
func (d *Duration) UnmarshalCSV(s string) (err error) {
	d.Duration, err = time.ParseDuration(s)
	return err
}
