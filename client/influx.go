package client

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/deltacat/dbstress/utils"
	"github.com/valyala/fasthttp"
)

type client struct {
	url []byte

	cfg InfluxConfig

	httpClient *fasthttp.Client
}

// NewInfluxClient return new influx (db/file) client instance
func NewInfluxClient(cfg InfluxConfig, dump string) (Client, error) {
	if dump != "" {
		return NewInfluxFileClient(dump, cfg)
	}
	return NewInfluxDbClient(cfg)
}

// NewInfluxDbClient create a new influxdb client instance
func NewInfluxDbClient(cfg InfluxConfig) (Client, error) {
	var httpClient *fasthttp.Client
	if cfg.TLSSkipVerify {
		httpClient = &fasthttp.Client{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return &client{
		url:        []byte(writeURLFromConfig(cfg)),
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *client) Create(command string) error {
	if command == "" {
		command = "CREATE DATABASE " + c.cfg.Database
	}

	vals := url.Values{}
	vals.Set("q", command)
	u, err := url.Parse(c.cfg.URL)
	if err != nil {
		return err
	}
	if c.cfg.User != "" && c.cfg.Pass != "" {
		u.User = url.UserPassword(c.cfg.User, c.cfg.Pass)
	}
	resp, err := http.PostForm(u.String()+"/query", vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(
			"Bad status code during Create(%s): %d, body: %s",
			command, resp.StatusCode, string(body),
		)
	}

	return nil
}

func (c *client) Send(b []byte) (latNs int64, statusCode int, body string, err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetContentTypeBytes([]byte("text/plain"))
	req.Header.SetMethodBytes([]byte("POST"))
	req.Header.SetRequestURIBytes(c.url)
	if c.cfg.Gzip != 0 {
		req.Header.SetBytesKV([]byte("Content-Encoding"), []byte("gzip"))
	}
	req.Header.SetContentLength(len(b))
	req.SetBody(b)

	resp := fasthttp.AcquireResponse()
	start := time.Now()

	do := fasthttp.Do
	if c.httpClient != nil {
		do = c.httpClient.Do
	}

	err = do(req, resp)
	latNs = time.Since(start).Nanoseconds()
	statusCode = resp.StatusCode()

	// Save the body.
	if statusCode != http.StatusNoContent {
		body = string(resp.Body())
	}

	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)

	return
}

func (c *client) SendString(string) (latNs int64, statusCode int, body string, err error) {
	return 0, 0, "", utils.ErrNotSupport
}

func (c *client) Close() error {
	// Nothing to do.
	return nil
}

func (c *client) Reset() error {
	return utils.ErrNotImplemented
}

func (c *client) Name() string {
	return c.cfg.Name
}

func (c *client) GzipLevel() int {
	return c.cfg.Gzip
}

type influxFileClient struct {
	database string

	mu    sync.Mutex
	f     *os.File
	batch uint
}

// NewInfluxFileClient create a file client for influxdb
func NewInfluxFileClient(path string, cfg InfluxConfig) (Client, error) {
	c := &influxFileClient{}

	var err error
	c.f, err = os.Create(path)
	if err != nil {
		return nil, err
	}

	if _, err := c.f.WriteString("# " + writeURLFromConfig(cfg) + "\n"); err != nil {
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

func (c *influxFileClient) GzipLevel() int {
	return 0
}

func writeURLFromConfig(cfg InfluxConfig) string {
	params := url.Values{}
	params.Set("db", cfg.Database)
	if cfg.User != "" {
		params.Set("u", cfg.User)
	}
	if cfg.Pass != "" {
		params.Set("p", cfg.Pass)
	}
	if cfg.RetentionPolicy != "" {
		params.Set("rp", cfg.RetentionPolicy)
	}
	if cfg.Precision != "n" && cfg.Precision != "" {
		params.Set("precision", cfg.Precision)
	}
	if cfg.Consistency != "one" && cfg.Consistency != "" {
		params.Set("consistency", cfg.Consistency)
	}

	return cfg.URL + "/write?" + params.Encode()
}
