package client

import (
	"net/http"
	"net/url"

	"encoding/json"

	"github.com/valyala/fasthttp"
)

type influxClientV2 struct {
	influxClient
	orgID  string
	bucket string
}

type dataMap map[string]interface{}
type queryMap map[string]string
type queryBucketsResult struct {
	Buckets []struct {
		Name string
		ID   string
	}
}

func (c *influxClientV2) Create(string) error {
	pl := map[string]interface{}{
		"orgID":          c.orgID,
		"name":           c.bucket,
		"retentionRules": []interface{}{},
	}
	_, err := c.sendCmd("/api/v2/buckets", "POST", nil, pl)
	return err
}

func (c *influxClientV2) Reset() error {
	resultRaw, err := c.sendCmd("/api/v2/buckets", "GET", queryMap{
		"orgID": c.orgID,
		"name":  c.bucket,
	}, nil)
	if err != nil {
		return err
	}
	bucketID := ""
	results := queryBucketsResult{}
	err = json.Unmarshal(resultRaw, &results)
	if err != nil {
		return err
	}
	for _, v := range results.Buckets {
		if v.Name == c.bucket {
			bucketID = v.ID
		}
	}

	_, err = c.sendCmd("/api/v2/buckets/"+bucketID, "DELETE", nil, nil)
	return err
}

func (c *influxClientV2) sendCmd(endpoint, methods string, query queryMap, payload dataMap) (result []byte, err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.Header.SetMethodBytes([]byte(methods))
	req.Header.SetRequestURIBytes([]byte(c.baseURL + "/" + endpoint))
	if c.token != "" {
		req.Header.Add("Authorization", "Token "+c.token)
	}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		req.Header.SetContentLength(len(b))
		req.SetBody(b)
	}

	if query != nil {
		for k, v := range query {
			req.URI().QueryArgs().Add(k, v)
		}
	}

	resp := fasthttp.AcquireResponse()

	do := fasthttp.Do
	if c.httpClient != nil {
		do = c.httpClient.Do
	}

	err = do(req, resp)

	// Save the body.
	if resp.StatusCode() != http.StatusNoContent {
		result = resp.Body()
	}

	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)

	return
}

func newInfluxDbV2Client(cfg InfluxConfig, httpClient *fasthttp.Client) *influxClientV2 {
	return &influxClientV2{
		influxClient: influxClient{
			name:       cfg.Name,
			baseURL:    cfg.URL,
			gzip:       cfg.Gzip,
			token:      cfg.V2.Token,
			httpClient: httpClient,
			writeURL:   []byte(writeURLFromConfigV2(cfg)),
		},
		orgID:  cfg.V2.OrgID,
		bucket: cfg.V2.Bucket,
	}
}

func writeURLFromConfigV2(cfg InfluxConfig) string {
	v2 := cfg.V2
	params := url.Values{}
	params.Set("org", v2.OrgID)
	if v2.Bucket != "" {
		params.Set("bucket", v2.Bucket)
	}
	if cfg.Precision != "n" && cfg.Precision != "" {
		params.Set("precision", cfg.Precision)
	}
	if cfg.Consistency != "one" && cfg.Consistency != "" {
		params.Set("consistency", cfg.Consistency)
	}

	return cfg.URL + "/api/v2/write?" + params.Encode()
}
