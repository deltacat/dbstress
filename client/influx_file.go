package client

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/deltacat/dbstress/utils"
)

type influxFileClient struct {
	database string

	mu    sync.Mutex
	f     *os.File
	path  string
	batch uint
}

// newInfluxFileClient create a file client for influxdb
func newInfluxFileClient(path string, cfg InfluxConfig) (Client, error) {
	c := &influxFileClient{
		path: path,
	}

	var err error
	c.f, err = os.Create(path)
	if err != nil {
		return nil, err
	}

	if _, err := c.f.WriteString("# " + writeURLFromConfigV1(cfg) + "\n"); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *influxFileClient) Create(command string) error {
	if command == "" {
		command = "CREATE DATABASE " + c.database
	}
	c.mu.Lock()
	_, err := fmt.Fprintf(c.f, "# create: %s\n\n", command)
	c.mu.Unlock()
	return err
}

func (c *influxFileClient) Send(b []byte) (latNs int64, statusCode int, body string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	statusCode = -1
	start := time.Now()
	defer func() {
		latNs = time.Since(start).Nanoseconds()
	}()

	c.batch++
	if _, err = fmt.Fprintf(c.f, "# Batch %d:\n", c.batch); err != nil {
		return
	}

	if _, err = c.f.Write(b); err != nil {
		return
	}

	if _, err = c.f.Write([]byte{'\n'}); err != nil {
		return
	}

	statusCode = 204
	return
}

func (c *influxFileClient) SendString(string) (latNs int64, statusCode int, body string, err error) {
	return 0, 0, "", utils.ErrNotSupport
}

func (c *influxFileClient) Close() error {
	return c.f.Close()
}

func (c *influxFileClient) Reset() error {
	return utils.ErrNotImplemented
}

func (c *influxFileClient) Name() string {
	return "InfluxFile"
}

func (c *influxFileClient) Connection() string {
	return c.path
}

func (c *influxFileClient) GzipLevel() int {
	return 0
}
