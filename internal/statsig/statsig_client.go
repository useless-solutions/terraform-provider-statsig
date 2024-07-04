package statsig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	Ctx      context.Context
	HostURL  string
	APIKey   string
	Metadata statsigMetadata
	Client   *http.Client
}

func NewClient(_ context.Context, apiKey string) (*Client, error) {
	return &Client{
		HostURL:  "https://api.statsig.com/console/v1",
		APIKey:   apiKey,
		Metadata: getStatsigMetadata(),
		Client:   &http.Client{Timeout: time.Second * 10},
	}, nil
}

// All API calls must include 'STATSIG-API-KEY' in the header. This is the apiKey value
func (c *Client) Get(method string, endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest(method, endpoint, nil, queryParams)
}

func (c *Client) doRequest(method string, endpoint string, requestBody interface{}, queryParams map[string]string) ([]byte, error) {
	req, err := c.buildRequest(method, endpoint, requestBody, queryParams)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Failed %s request to %s with status code %d.", method, req.URL, res.StatusCode)
	}
	defer res.Body.Close()

	parsedBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, parsedBody)
	}

	return parsedBody, err
}

func (c *Client) buildRequest(method, endpoint string, body interface{}, queryParams map[string]string) (*http.Request, error) {
	var bodyBuf io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyBuf = bytes.NewBuffer(bodyBytes)
	} else {
		if method == "POST" {
			bodyBuf = bytes.NewBufferString("{}")
		}
	}
	url := fmt.Sprintf("%s/%s", c.HostURL, endpoint)
	req, err := http.NewRequest(method, url, bodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("STATSIG-API-KEY", c.APIKey)
	if method == "POST" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	req.Header.Add("STATSIG-SDK-TYPE", c.Metadata.SDKType)
	req.Header.Add("STATSIG-SDK-VERSION", c.Metadata.SDKVersion)

	// Add query parameters if any
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}