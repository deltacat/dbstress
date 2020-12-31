package stress

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/deltacat/dbstress/client"
)

// Sink sink interface
type Sink interface {
	Chan() chan WriteResult
	Open()
	Close()
}

// ErrorSink sink interface implementation for errors
type ErrorSink struct {
	Ch     chan WriteResult
	strict bool

	wg sync.WaitGroup
}

// NewErrorSink create a new error sink
func NewErrorSink(nWriters int) *ErrorSink {
	s := &ErrorSink{
		Ch: make(chan WriteResult, 8*nWriters),
	}

	s.wg.Add(1)
	go s.checkErrors()

	return s
}

// Open open sink
func (s *ErrorSink) Open() {
	s.wg.Add(1)
	go s.checkErrors()
}

// Close close sink
func (s *ErrorSink) Close() {
	close(s.Ch)
	s.wg.Wait()
}

func (s *ErrorSink) checkErrors() {
	defer s.wg.Done()

	const timeFormat = "[2006-01-02 15:04:05]"
	for r := range s.Ch {
		if r.Err != nil {
			fmt.Fprintln(os.Stderr, time.Now().Format(timeFormat), "Error sending write:", r.Err.Error())
			continue
		}

		if r.StatusCode != 204 {
			fmt.Fprintln(os.Stderr, time.Now().Format(timeFormat), "Unexpected write: status", r.StatusCode, ", body:", r.Body)
		}

		// If we're running in strict mode then we give up at the first error.
		if s.strict && (r.Err != nil || r.StatusCode != 204) {
			os.Exit(1)
		}
	}
}

// Chan return sink chan
func (s *ErrorSink) Chan() chan WriteResult {
	return s.Ch
}

// MultiSink sink interface implementation for multi
type MultiSink struct {
	sinks []Sink
	Ch    chan WriteResult
	wg    sync.WaitGroup
	open  bool
}

// NewMultiSink create a new multi sink
func NewMultiSink(nWriters int) *MultiSink {
	return &MultiSink{
		Ch: make(chan WriteResult, 8*nWriters),
	}
}

// Chan return chan
func (s *MultiSink) Chan() chan WriteResult {
	return s.Ch
}

// Open open sink
func (s *MultiSink) Open() {
	s.open = true
	for _, sink := range s.sinks {
		sink.Open()
	}
	s.wg.Add(1)
	go s.run()
}

func (s *MultiSink) run() {
	defer s.wg.Done()
	const timeFormat = "[2006-01-02 15:04:05]"
	for r := range s.Ch {
		for _, sink := range s.sinks {
			select {
			case sink.Chan() <- r:
			default:
				fmt.Fprintln(os.Stderr, time.Now().Format(timeFormat), "Failed to send to sin")
			}
		}
	}
}

// Close close sink
func (s *MultiSink) Close() {
	close(s.Ch)
	s.wg.Wait()
	s.open = false
	for _, sink := range s.sinks {
		sink.Close()
	}
}

// AddSink add sink
func (s *MultiSink) AddSink(sink Sink) error {
	if s.open {
		return errors.New("Cannot add sink to open multiSink")
	}

	s.sinks = append(s.sinks, sink)

	return nil
}

// InfluxDBSink implement sink interface
type InfluxDBSink struct {
	Ch     chan WriteResult
	client client.Client
	buf    *bytes.Buffer
	ticker *time.Ticker
}

// NewInfluxDBSink create a new InfluxDBSink instance
func NewInfluxDBSink(nWriters int, url, db string) *InfluxDBSink {
	cfg := client.InfluxConfig{
		URL:         url,
		APIVersion:  1,
		Precision:   "ns",
		Consistency: "any",
		Gzip:        0,
	}
	cfg.V1.Database = db
	cfg.V1.RetentionPolicy = "autogen"

	cli, _ := client.NewInfluxClient(cfg, "")

	return &InfluxDBSink{
		Ch:     make(chan WriteResult, 8*nWriters),
		client: cli,
		buf:    bytes.NewBuffer(nil),
	}
}

// Chan return the sink chan
func (s *InfluxDBSink) Chan() chan WriteResult {
	return s.Ch
}

// Open open the influxdb sink
func (s *InfluxDBSink) Open() {
	s.ticker = time.NewTicker(time.Second)
	err := s.client.Create("")
	if err != nil {
		panic(err)
	}

	go s.run()
}

// Close close influxdb sink
func (s *InfluxDBSink) Close() {
}

func (s *InfluxDBSink) run() {
	for {
		select {
		case <-s.ticker.C:
			// Write batch
			s.client.Send(s.buf.Bytes())
			s.buf.Reset()
		case result := <-s.Ch:
			// Add to batch
			if result.Err != nil {
				continue
			}
			s.buf.WriteString(fmt.Sprintf("req,status=%v latNs=%v %v\n", result.StatusCode, result.LatNs, result.Timestamp))
		}
	}
}
