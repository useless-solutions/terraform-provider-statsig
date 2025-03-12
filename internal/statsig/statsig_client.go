package statsig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	client "github.com/useless-solutions/statsig-go-client"
)

type Client struct {
	Ctx      context.Context
	HostURL  string
	APIKey   string
	Metadata statsigMetadata
	Client   *http.Client
}

// ErrorResponse is the representation of the response body when an error occurs. This is different from
// the APIResponse struct as it only contains the message and status code of the error, rather than the data.
type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status"`
}

// NewClient creates a new Statsig client with the provided API key.
// The Client instance includes an HTTP client with a 10-second timeout to be used for API requests.
func NewClient(_ context.Context, apiKey string) (*Client, error) {
	return &Client{
		HostURL:  "https://statsigapi.net/console/v1",
		APIKey:   apiKey,
		Metadata: getStatsigMetadata(),
		Client:   &http.Client{Timeout: time.Second * 10},
	}, nil
}

func NewAndImprovedClient(apiKey string) client.Client {
	metadata := getStatsigMetadata()
	return client.Client{
		Server: "https://statsigapi.net/console/v1",
		Client: &http.Client{Timeout: time.Second * 10},
		RequestEditors: []client.RequestEditorFn{
			func(ctx context.Context, req *http.Request) error {
				req.Header.Add("STATSIG-API-KEY", apiKey)
				if req.Method == "POST" || req.Method == "PATCH" {
					req.Header.Set("Content-Type", "application/json; charset=UTF-8")
				}
				req.Header.Add("STATSIG-SDK-TYPE", metadata.SDKType)
				req.Header.Add("STATSIG-SDK-VERSION", metadata.SDKVersion)
				return nil
			},
		},
	}
}

// Get performs a GET request with the provided endpoint and queryParams.
// There is no request body.
//
// Get returns the response body as a byte slice, or an error if the request fails.
func (c *Client) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("GET", endpoint, nil, queryParams)
}

// Post performs a POST request with the provided endpoint and requestBody.
// The request body is marshalled into JSON before being sent.
//
// Post returns the response body as a byte slice, or an error if the request fails.
func (c *Client) Post(endpoint string, requestBody interface{}) ([]byte, error) {
	return c.doRequest("POST", endpoint, requestBody, nil)
}

// Patch performs a PATCH request with the provided endpoint and requestBody.
// The request body is marshalled into JSON before being sent.
//
// Patch returns the response body as a byte slice, or an error if the request fails.
func (c *Client) Patch(endpoint string, requestBody interface{}) ([]byte, error) {
	return c.doRequest("PATCH", endpoint, requestBody, nil)
}

// Delete performs a DELETE request with the provided endpoint and queryParams.
// There is no request body.
//
// Delete returns the response body as a byte slice, or an error if the request fails.
func (c *Client) Delete(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("DELETE", endpoint, nil, queryParams)
}

// doRequest performs an HTTP request that is built with the provided method, endpoint, body, and queryParams.
// The request is first build using the buildRequest method, and then executed using the Statsig Client's HTTP client.
//
// The API returns an error message in the response body when an error occurs. Unknown (unexpected) errors are parsed
// and returned as-is, while known errors are returned as a formatted error message.
func (c *Client) doRequest(method string, endpoint string, requestBody interface{}, queryParams map[string]string) ([]byte, error) {
	req, err := c.buildRequest(method, endpoint, requestBody, queryParams)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch {
	case res.StatusCode == 401:
		return nil, fmt.Errorf("Unauthorized request to %s. Please check your API key.", req.URL)
	case res.StatusCode < 200 || res.StatusCode >= 300:
		parsedBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		errorResponse := ErrorResponse{}
		fmt.Sprintln(string(parsedBody))
		if err := json.Unmarshal(parsedBody, &errorResponse); err != nil {
			return nil, fmt.Errorf("Failed to perform request to %s with status code %d and response body: %s", req.URL, res.StatusCode, errorResponse.Message)
		}

		return nil, fmt.Errorf(errorResponse.Message)
	}

	parsedBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return parsedBody, err
}

// buildRequest creates an HTTP request with the provided method, endpoint, body, and query parameters.
// The request includes Statsig-specific headers and metadata, such as the SDK type and version.
//
// Uniquely, the API Key is included in a custom STATSIG-API-KEY header, rather than the standard Authorization header.
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
