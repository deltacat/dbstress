package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/lineprotocol"
	"github.com/deltacat/dbstress/point"
	"github.com/deltacat/dbstress/stress"
	"github.com/deltacat/dbstress/write"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statsHost, statsDB      string
	createCommand, dump     string
	seriesN                 int
	batchSize, pointsN, pps uint64
	runtime                 time.Duration
	tick                    time.Duration
	fast, quiet             bool
	strict, kapacitorMode   bool
	recordStats             bool
	tlsSkipVerify           bool
)

var insertCmd = &cobra.Command{
	Use:   "insert SERIES FIELDS",
	Short: "Insert data into DB",
	Long:  "",
	Run:   runInsert,
}

func runInsert(cmd *cobra.Command, args []string) {
	insert(config.Cfg)
}

func insert(cfg config.Config) {
	seriesKey := cfg.Points.SeriesKey
	fieldStr := cfg.Points.FieldsStr
	if !strings.Contains(seriesKey, ",") && !strings.Contains(seriesKey, "=") {
		logrus.Warnf("expect series like 'ctr,some=tag', got '%s'", seriesKey)
		os.Exit(1)
		return
	}

	concurrency := pps / batchSize
	// PPS takes precedence over batchSize.
	// Adjust accordingly.
	if pps < batchSize {
		batchSize = pps
		concurrency = 1
	}
	if !quiet {
		fmt.Printf("Using point template: %s %s <timestamp>\n", seriesKey, fieldStr)
		fmt.Printf("Using batch size of %d line(s)\n", batchSize)
		fmt.Printf("Spreading writes across %d series\n", seriesN)
		if fast {
			fmt.Println("Output is unthrottled")
		} else {
			fmt.Printf("Throttling output to ~%d points/sec\n", pps)
		}
		fmt.Printf("Using %d concurrent writer(s)\n", concurrency)

		fmt.Printf("Running until ~%d points sent or until ~%v has elapsed\n", pointsN, runtime)
	}

	c := client()

	if !kapacitorMode {
		if err := c.Create(createCommand); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create database:", err.Error())
			fmt.Fprintln(os.Stderr, "Aborting.")
			os.Exit(1)
			return
		}
	}

	pts := point.NewPoints(seriesKey, fieldStr, seriesN, lineprotocol.Nanosecond)

	startSplit := 0
	inc := int(seriesN) / int(concurrency)
	endSplit := inc

	sink := newMultiSink(int(concurrency))
	sink.AddSink(newErrorSink(int(concurrency)))

	if recordStats {
		sink.AddSink(newInfluxDBSink(int(concurrency), statsHost, statsDB))
	}

	sink.Open()

	var wg sync.WaitGroup
	wg.Add(int(concurrency))

	var totalWritten uint64

	start := time.Now()
	gzip := cfg.Connection.Influxdb.Gzip
	for i := uint64(0); i < concurrency; i++ {

		go func(startSplit, endSplit int) {
			tick := time.Tick(tick)

			if fast {
				tick = time.Tick(time.Nanosecond)
			}

			cfg := stress.WriteConfig{
				BatchSize: batchSize,
				MaxPoints: pointsN / concurrency, // divide by concurreny
				GzipLevel: gzip,
				Deadline:  time.Now().Add(runtime),
				Tick:      tick,
				Results:   sink.Chan(),
			}

			// Ignore duration from a single call to Write.
			pointsWritten, _ := stress.Write(pts[startSplit:endSplit], c, cfg)
			atomic.AddUint64(&totalWritten, pointsWritten)

			wg.Done()
		}(startSplit, endSplit)

		startSplit = endSplit
		endSplit += inc
	}

	wg.Wait()
	totalTime := time.Since(start)
	if err := c.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing client: %v\n", err.Error())
	}

	sink.Close()
	throughput := int(float64(totalWritten) / totalTime.Seconds())
	if quiet {
		fmt.Println(throughput)
	} else {
		fmt.Println("Write Throughput:", throughput)
		fmt.Println("Points Written:", totalWritten)
	}
}

func init() {
	rootCmd.AddCommand(insertCmd)

	insertCmd.Flags().StringVarP(&statsHost, "stats-host", "", "http://localhost:8086", "Address of InfluxDB instance where runtime statistics will be recorded")
	insertCmd.Flags().StringVarP(&statsDB, "stats-db", "", "stress_stats", "Database that statistics will be written to")
	insertCmd.Flags().BoolVarP(&recordStats, "stats", "", false, "Record runtime statistics")

	insertCmd.Flags().IntVarP(&seriesN, "series", "s", 100000, "number of series that will be written")
	insertCmd.Flags().Uint64VarP(&pointsN, "points", "n", math.MaxUint64, "number of points that will be written")
	insertCmd.Flags().Uint64VarP(&batchSize, "batch-size", "b", 10000, "number of points in a batch")
	insertCmd.Flags().Uint64VarP(&pps, "pps", "", 200000, "Points Per Second")
	insertCmd.Flags().DurationVarP(&runtime, "runtime", "r", time.Duration(math.MaxInt64), "Total time that the test will run")
	insertCmd.Flags().DurationVarP(&tick, "tick", "", time.Second, "Amount of time between request")
	insertCmd.Flags().BoolVarP(&fast, "fast", "f", false, "Run as fast as possible")
	insertCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only print the write throughput")
	insertCmd.Flags().StringVar(&createCommand, "create", "", "Use a custom create database command")
	insertCmd.Flags().BoolVarP(&kapacitorMode, "kapacitor", "k", false, "Use Kapacitor mode, namely do not try to run any queries.")
	insertCmd.Flags().StringVar(&dump, "dump", "", "Dump to given file instead of writing over HTTP")
	insertCmd.Flags().BoolVarP(&strict, "strict", "", false, "Strict mode will exit as soon as an error or unexpected status is encountered")
}

func client() write.Client {
	influxCfg := config.Cfg.Connection.Influxdb
	if dump != "" {
		c, err := write.NewFileClient(dump, influxCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening file:", err)
			os.Exit(1)
			return c
		}

		return c
	}
	return write.NewClient(influxCfg)
}

// Sink sink interface
type Sink interface {
	Chan() chan stress.WriteResult
	Open()
	Close()
}

type errorSink struct {
	Ch chan stress.WriteResult

	wg sync.WaitGroup
}

func newErrorSink(nWriters int) *errorSink {
	s := &errorSink{
		Ch: make(chan stress.WriteResult, 8*nWriters),
	}

	s.wg.Add(1)
	go s.checkErrors()

	return s
}

func (s *errorSink) Open() {
	s.wg.Add(1)
	go s.checkErrors()
}

func (s *errorSink) Close() {
	close(s.Ch)
	s.wg.Wait()
}

func (s *errorSink) checkErrors() {
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
		if strict && (r.Err != nil || r.StatusCode != 204) {
			os.Exit(1)
		}
	}
}

func (s *errorSink) Chan() chan stress.WriteResult {
	return s.Ch
}

type multiSink struct {
	Ch chan stress.WriteResult

	sinks []Sink

	open bool
}

func newMultiSink(nWriters int) *multiSink {
	return &multiSink{
		Ch: make(chan stress.WriteResult, 8*nWriters),
	}
}

func (s *multiSink) Chan() chan stress.WriteResult {
	return s.Ch
}

func (s *multiSink) Open() {
	s.open = true
	for _, sink := range s.sinks {
		sink.Open()
	}
	go s.run()
}

func (s *multiSink) run() {
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

func (s *multiSink) Close() {
	s.open = false
	for _, sink := range s.sinks {
		sink.Close()
	}
}

func (s *multiSink) AddSink(sink Sink) error {
	if s.open {
		return errors.New("Cannot add sink to open multiSink")
	}

	s.sinks = append(s.sinks, sink)

	return nil
}

type influxDBSink struct {
	Ch     chan stress.WriteResult
	client write.Client
	buf    *bytes.Buffer
	ticker *time.Ticker
}

func newInfluxDBSink(nWriters int, url, db string) *influxDBSink {
	cfg := write.ClientConfig{
		URL:             url,
		Database:        db,
		RetentionPolicy: "autogen",
		Precision:       "ns",
		Consistency:     "any",
		Gzip:            0,
	}

	return &influxDBSink{
		Ch:     make(chan stress.WriteResult, 8*nWriters),
		client: write.NewClient(cfg),
		buf:    bytes.NewBuffer(nil),
	}
}

func (s *influxDBSink) Chan() chan stress.WriteResult {
	return s.Ch
}

func (s *influxDBSink) Open() {
	s.ticker = time.NewTicker(time.Second)
	err := s.client.Create("")
	if err != nil {
		panic(err)
	}

	go s.run()
}

func (s *influxDBSink) Close() {
}

func (s *influxDBSink) run() {
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
