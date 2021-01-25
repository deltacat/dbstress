package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

type influxClientV1 struct {
	influxClient
	database string
	user     string
	pass     string
}

func newInfluxDbV1Client(cfg InfluxConfig, httpClient *fasthttp.Client) *influxClientV1 {
	return &influxClientV1{
		influxClient: influxClient{
			name:       cfg.Name,
			baseURL:    cfg.URL,
			httpClient: httpClient,
			writeURL:   []byte(writeURLFromConfigV1(cfg)),
		},
		database: cfg.V1.Database,
		user:     cfg.V1.User,
		pass:     cfg.V1.Pass,
	}
}

func (c *influxClientV1) Create(command string) error {
	if command == "" {
		command = "CREATE DATABASE " + c.database
	}
	return c.sendCmd(command)
}

func (c *influxClientV1) Reset() error {
	return c.sendCmd("DROP DATABASE " + c.database)
}

func (c *influxClientV1) sendCmd(cmd string) error {
	vals := url.Values{}
	vals.Set("q", cmd)

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	if c.user != "" && c.pass != "" {
		u.User = url.UserPassword(c.user, c.pass)
	}

	req, err := http.NewRequest("POST", u.String()+"/query", strings.NewReader(vals.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(vals.Encode())))
	req.PostForm = vals

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(
			"Bad status code during execute cmd (%s): %d, body: %s",
			cmd, resp.StatusCode, string(body),
		)
	}

	return nil
}

func writeURLFromConfigV1(cfg InfluxConfig) string {
	params := url.Values{}
	v1 := cfg.V1
	params.Set("db", v1.Database)
	if v1.User != "" {
		params.Set("u", v1.User)
	}
	if cfg.V1.Pass != "" {
		params.Set("p", v1.Pass)
	}
	if v1.RetentionPolicy != "" {
		params.Set("rp", v1.RetentionPolicy)
	}
	if cfg.Precision != "n" && cfg.Precision != "" {
		params.Set("precision", cfg.Precision)
	}
	if cfg.Consistency != "one" && cfg.Consistency != "" {
		params.Set("consistency", cfg.Consistency)
	}

	return cfg.URL + "/write?" + params.Encode()
}
