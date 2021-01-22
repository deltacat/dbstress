package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/deltacat/dbstress/utils"
	"github.com/valyala/fasthttp"
)

type influxClient struct {
	// preset fields
	name    string
	baseURL string
	token   string
	gzip    int

	// built out fields
	httpClient *fasthttp.Client
	writeURL   []byte
}

// NewInfluxClient return new influx (db/file) client instance
func NewInfluxClient(cfg InfluxConfig, dump string) (Client, error) {
	// return file client
	if dump != "" {
		return newInfluxFileClient(dump, cfg)
	}

	var httpClient *fasthttp.Client
	if cfg.TLSSkipVerify {
		httpClient = &fasthttp.Client{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	if err := checkHealth(cfg.URL); err != nil {
		return nil, err
	}

	// return v2 client
	if cfg.APIVersion == 2 {
		return newInfluxDbV2Client(cfg, httpClient), nil
	}

	// return v1 client
	return newInfluxDbV1Client(cfg, httpClient), nil
}

func checkHealth(host string) error {
	resp, err := http.Get(host + "/health")
	if err != nil {
		return errors.Unwrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(
			"check host health failed, status code: %d, body: %s",
			resp.StatusCode, string(body),
		)
	}

	return nil
}

func (c *influxClient) Send(b []byte) (latNs int64, statusCode int, body string, err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetContentTypeBytes([]byte("text/plain"))
	req.Header.SetMethodBytes([]byte("POST"))
	req.Header.SetRequestURIBytes(c.writeURL)
	if c.token != "" {
		req.Header.Add("Authorization", "Token "+c.token)
	}
	if c.gzip != 0 {
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
	if statusCode >= http.StatusBadRequest {
		err = errors.New(http.StatusText(statusCode))
	}

	// Save the body.
	if statusCode != http.StatusNoContent {
		body = string(resp.Body())
	}

	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)

	return
}

func (c *influxClient) SendString(string) (latNs int64, statusCode int, body string, err error) {
	return 0, 0, "", utils.ErrNotSupport
}

func (c *influxClient) Close() error {
	// Nothing to do.
	return nil
}

func (c *influxClient) Name() string {
	return c.name
}

func (c *influxClient) Connection() string {
	ps := strings.Split(c.baseURL, "//")
	return ps[len(ps)-1]
}

func (c *influxClient) GzipLevel() int {
	return c.gzip
}
