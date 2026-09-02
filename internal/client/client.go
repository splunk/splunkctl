package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/splunk/splunkctl/internal/config"
)

type Request struct {
	Method string
	Path   string
	ID     string
	Params map[string]string
	Output string // "json" | "table" | "text" — passed to output renderer, not used by client itself
}

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

// HTTPClient returns the underlying http.Client for callers that need raw HTTP access.
func (c *Client) HTTPClient() *http.Client { return c.httpClient }

// Host returns the base Splunk URL (e.g. https://host:8089) without trailing slash.
func (c *Client) Host() string { return strings.TrimRight(c.cfg.Host, "/") }

// Token returns the configured bearer token.
func (c *Client) Token() string { return c.cfg.Token }

// RawDoJSON performs a POST/PUT to an absolute path with a JSON body.
// Used for KV Store data operations which require Content-Type: application/json.
func (c *Client) RawDoJSON(ctx context.Context, method, path string, payload any) ([]byte, int, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	fullURL := c.Host() + path

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// RawDoAbs performs an HTTP request to an absolute path (no /services prefix added).
// path should start with / and be relative to the host root (e.g. /servicesNS/nobody/search/...).
func (c *Client) RawDoAbs(ctx context.Context, method, path string, body url.Values) ([]byte, int, error) {
	fullURL := c.Host() + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(body.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	q := req.URL.Query()
	q.Set("output_mode", "json")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func New(cfg *config.Config) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure},
	}
	return &Client{cfg: cfg, httpClient: &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}}
}

func (c *Client) Do(ctx context.Context, req Request) ([]map[string]any, error) {
	path := req.Path
	if req.ID != "" {
		path = strings.TrimRight(path, "/") + "/" + url.PathEscape(req.ID)
	}

	fullURL := strings.TrimRight(c.cfg.Host, "/") + "/services" + path

	var httpReq *http.Request
	var err error

	switch req.Method {
	case "GET", "DELETE":
		httpReq, err = http.NewRequestWithContext(ctx, req.Method, fullURL, nil)
		if err != nil {
			return nil, err
		}
		q := httpReq.URL.Query()
		q.Set("output_mode", "json")
		for k, v := range req.Params {
			q.Set(k, v)
		}
		httpReq.URL.RawQuery = q.Encode()

	case "POST":
		form := url.Values{}
		for k, v := range req.Params {
			form.Set(k, v)
		}
		httpReq, err = http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		q := httpReq.URL.Query()
		q.Set("output_mode", "json")
		httpReq.URL.RawQuery = q.Encode()
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.Token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	return parseEntries(body)
}

// Upload performs a multipart form POST to upload a file (e.g. lookup CSV).
// fieldName is the form field name, fileName is the destination name, data is the file content.
func (c *Client) Upload(ctx context.Context, path, fieldName, fileName string, data []byte) ([]map[string]any, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("output_mode", "json")
	part, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	w.Close()

	fullURL := strings.TrimRight(c.cfg.Host, "/") + "/services" + path
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp.StatusCode, respBody)
	}
	return parseEntries(respBody)
}

// Describe fetches the _new introspection endpoint for path and returns the writable fields.
// Splunk exposes required/optional/wildcard field lists at <path>/_new.
func (c *Client) Describe(ctx context.Context, path string) (map[string]any, error) {
	fullURL := strings.TrimRight(c.cfg.Host, "/") + "/services" + strings.TrimRight(path, "/") + "/_new"
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("output_mode", "json")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	var feed struct {
		Entry []struct {
			Fields map[string]any `json:"fields"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil || len(feed.Entry) == 0 {
		return nil, fmt.Errorf("unexpected describe response")
	}
	return feed.Entry[0].Fields, nil
}

// RawDo performs a raw HTTP request and returns the response body bytes and status code.
// Callers are responsible for parsing the response. Used by override commands that need
// direct control over the response (e.g. search job polling).
func (c *Client) RawDo(ctx context.Context, method, path string, body url.Values) ([]byte, int, error) {
	fullURL := strings.TrimRight(c.cfg.Host, "/") + "/services" + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(body.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	q := req.URL.Query()
	q.Set("output_mode", "json")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// SplunkError is returned for 4xx/5xx HTTP responses from the Splunk REST API.
type SplunkError struct {
	Status  int
	Message string
	Code    string
}

func (e *SplunkError) Error() string {
	return fmt.Sprintf("splunk error %d (%s): %s", e.Status, e.Code, e.Message)
}

func parseErrorResponse(status int, body []byte) error {
	var feed struct {
		Messages []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"messages"`
	}
	code := httpStatusToCode(status)
	if json.Unmarshal(body, &feed) == nil && len(feed.Messages) > 0 {
		return &SplunkError{Status: status, Message: feed.Messages[0].Text, Code: code}
	}
	return &SplunkError{Status: status, Message: string(body), Code: code}
}

func httpStatusToCode(status int) string {
	switch status {
	case 401, 403:
		return "AUTH_ERROR"
	case 404:
		return "NOT_FOUND"
	case 429:
		return "RATE_LIMITED"
	default:
		return "SERVER_ERROR"
	}
}

func parseEntries(body []byte) ([]map[string]any, error) {
	var feed struct {
		Entry []struct {
			Name    string         `json:"name"`
			Content map[string]any `json:"content"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		var raw []map[string]any
		if err2 := json.Unmarshal(body, &raw); err2 == nil {
			return raw, nil
		}
		return nil, fmt.Errorf("cannot parse response: %w", err)
	}
	results := make([]map[string]any, 0, len(feed.Entry))
	for _, e := range feed.Entry {
		row := map[string]any{"name": e.Name}
		for k, v := range e.Content {
			row[k] = v
		}
		results = append(results, row)
	}
	return results, nil
}
