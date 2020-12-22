package stress

import (
	"bytes"
	"compress/gzip"
	"io"
	"time"

	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/data/influx/lineprotocol"
	"github.com/deltacat/dbstress/data/mysql"
)

// WriteResult contains the latency, status code, and error type
// each time a write happens.
type WriteResult struct {
	LatNs      int64
	StatusCode int
	Body       string // Only populated when unusual status code encountered.
	Err        error
	Timestamp  int64
}

// WriteConfig specifies the configuration for the Write function.
type WriteConfig struct {
	BatchSize uint64
	MaxPoints uint64

	// If 0 (NoCompression), do not gzip at all.
	// Otherwise, pass this value to the gzip writer.
	GzipLevel int

	Deadline time.Time
	Tick     <-chan time.Time
	Results  chan<- WriteResult
}

// WriteInflux takes in a slice of lineprotocol.Points, a write.Client, and a WriteConfig. It will attempt
// to write data to the target until one of the following conditions is met.
// 1. We reach that MaxPoints specified in the WriteConfig.
// 2. We've passed the Deadline specified in the WriteConfig.
func WriteInflux(pts []lineprotocol.Point, c client.Client, cfg WriteConfig) (uint64, time.Duration) {
	if cfg.Results == nil {
		panic("Results Channel on WriteConfig cannot be nil")
	}
	var pointCount uint64

	start := time.Now()
	buf := bytes.NewBuffer(nil)
	t := time.Now()

	var w io.Writer = buf

	doGzip := cfg.GzipLevel != 0
	var gzw *gzip.Writer
	if doGzip {
		var err error
		gzw, err = gzip.NewWriterLevel(w, cfg.GzipLevel)
		if err != nil {
			// Should only happen with an invalid gzip level?
			panic(err)
		}
		w = gzw
	}

	tPrev := t
WRITE_BATCHES:
	for {
		if t.After(cfg.Deadline) {
			break WRITE_BATCHES
		}

		if pointCount >= cfg.MaxPoints {
			break
		}

		for _, pt := range pts {
			pointCount++
			pt.SetTime(t)
			lineprotocol.WritePoint(w, pt)
			if pointCount%cfg.BatchSize == 0 {
				if doGzip {
					// Must Close, not Flush, to write full gzip content to underlying bytes buffer.
					if err := gzw.Close(); err != nil {
						panic(err)
					}
				}
				sendBatchInflux(c, buf, cfg.Results)
				if doGzip {
					// sendBatch already reset the bytes buffer.
					// Reset the gzip writer to start clean.
					gzw.Reset(buf)
				}

				t = <-cfg.Tick
				if t.After(cfg.Deadline) {
					break WRITE_BATCHES
				}

				if pointCount >= cfg.MaxPoints {
					break
				}

			}
			pt.Update()
		}

		// Avoid timestamp colision when batch size > pts
		if t.After(tPrev) {
			tPrev = t
			continue
		}
		t = t.Add(1 * time.Nanosecond)
	}

	return pointCount, time.Since(start)
}

func sendBatchInflux(c client.Client, buf *bytes.Buffer, ch chan<- WriteResult) {
	lat, status, body, err := c.Send(buf.Bytes())
	buf.Reset()
	select {
	case ch <- WriteResult{LatNs: lat, StatusCode: status, Body: body, Err: err, Timestamp: time.Now().UnixNano()}:
	default:
	}
}

// WriteMySQL writes rows into mysql.
// Simlar as influx processing, it will attempt to write data to the target until one of the following conditions is met.
// 1. We reach that MaxPoints specified in the WriteConfig.
// 2. We've passed the Deadline specified in the WriteConfig.
func WriteMySQL(table mysql.Table, c client.Client, cfg WriteConfig) (uint64, time.Duration) {
	if cfg.Results == nil {
		panic("Results Channel on WriteConfig cannot be nil")
	}
	var pointCount uint64

	start := time.Now()
	t := time.Now()

	tPrev := t
	for {
		if t.After(cfg.Deadline) || pointCount >= cfg.MaxPoints {
			break
		}
		pointCount += table.GetRowsNum()

		sendBatchMySQL(c, table.GenInsertStmt(), cfg.Results)

		t = <-cfg.Tick

		table.Update()

		// Avoid timestamp colision when batch size > pts
		if t.After(tPrev) {
			tPrev = t
			continue
		}
		t = t.Add(1 * time.Nanosecond)
	}

	return pointCount, time.Since(start)
}

func sendBatchMySQL(c client.Client, query string, ch chan<- WriteResult) {
	lat, status, body, err := c.SendString(query)
	select {
	case ch <- WriteResult{LatNs: lat, StatusCode: status, Body: body, Err: err, Timestamp: time.Now().UnixNano()}:
	default:
	}
}
