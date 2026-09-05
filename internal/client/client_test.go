package client_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/config"
)

func TestClient_InjectsBearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Host: srv.URL, Token: "mytoken", Insecure: true}
	c := client.New(cfg)
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer mytoken" {
		t.Errorf("expected Bearer mytoken, got %q", gotAuth)
	}
}

func TestClient_BuildsCorrectURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Host: srv.URL, Token: "tok", Insecure: true}
	c := client.New(cfg)
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/", ID: "myindex"})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/services/data/indexes/myindex" {
		t.Errorf("expected /services/data/indexes/myindex, got %q", gotPath)
	}
}

func TestClient_OutputModeJSON(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("output_mode") != "json" {
			t.Errorf("expected output_mode=json, got %q", r.URL.Query().Get("output_mode"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[{"name":"main","content":{"maxDataSizeMB":500}}]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Host: srv.URL, Token: "tok", Insecure: true}
	c := client.New(cfg)
	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "main" {
		t.Errorf("expected name=main, got %v", results[0]["name"])
	}
}

func TestClient_404ReturnsError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"index not found"}]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Host: srv.URL, Token: "tok", Insecure: true}
	c := client.New(cfg)
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/", ID: "missing"})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "NOT_FOUND") {
		t.Errorf("expected NOT_FOUND in error, got: %v", err)
	}
}

// RawDo must prepend /services to the path.
func TestRawDo_PrependsServicesPrefix(t *testing.T) {
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	c.RawDo(context.Background(), "GET", "/search/jobs", nil)
	if gotPath != "/services/search/jobs" {
		t.Errorf("expected /services/search/jobs, got %q", gotPath)
	}
}

// RawDo must always set output_mode=json in the query string.
func TestRawDo_SetsOutputModeJSON(t *testing.T) {
	var gotMode string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMode = r.URL.Query().Get("output_mode")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	c.RawDo(context.Background(), "GET", "/search/jobs", nil)
	if gotMode != "json" {
		t.Errorf("expected output_mode=json, got %q", gotMode)
	}
}

// RawDo must send form body when params are provided.
func TestRawDo_SendsFormBody(t *testing.T) {
	var gotBody string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	body := url.Values{}
	body.Set("search", "index=main")
	c.RawDo(context.Background(), "POST", "/search/jobs", body)
	if !strings.Contains(gotBody, "search=") {
		t.Errorf("expected form body with search=, got %q", gotBody)
	}
}

// Host must strip trailing slash so URLs do not get a double slash.
func TestHost_StripsTrailingSlash(t *testing.T) {
	c := client.New(&config.Config{Host: "https://myhost:8089/", Token: "tok"})
	got := c.Host()
	if got != "https://myhost:8089" {
		t.Errorf("expected https://myhost:8089, got %q", got)
	}
}

// Host must return the URL unchanged when there is no trailing slash.
func TestHost_NoTrailingSlash(t *testing.T) {
	c := client.New(&config.Config{Host: "https://myhost:8089", Token: "tok"})
	got := c.Host()
	if got != "https://myhost:8089" {
		t.Errorf("expected https://myhost:8089, got %q", got)
	}
}

// Upload must send a multipart/form-data request with the file content.
func TestUpload_SendsMultipartBody(t *testing.T) {
	var gotContentType string
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Upload(context.Background(), "/data/lookup-table-files/", "file", "test.csv", []byte("a,b\n1,2"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content-type, got %q", gotContentType)
	}
	if !strings.Contains(string(gotBody), "test.csv") {
		t.Errorf("expected filename test.csv in body, got %q", string(gotBody))
	}
	if !strings.Contains(string(gotBody), "a,b") {
		t.Errorf("expected file content in body, got %q", string(gotBody))
	}
}

func TestClient_RawDoJSON_SendsJSONBody(t *testing.T) {
	var gotContentType string
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, status, err := c.RawDoJSON(context.Background(), "POST", "/servicesNS/nobody/search/storage/collections/data/test/", map[string]any{"name": "main"})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected application/json content-type, got %q", gotContentType)
	}
	if !strings.Contains(string(gotBody), "main") {
		t.Errorf("expected JSON body to contain 'main', got %q", string(gotBody))
	}
}

func TestClient_WriteCheckUsesHTTPMethod(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			requests := 0
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
			}))
			defer srv.Close()

			checked := 0
			denied := errors.New("denied")
			c := client.NewWithWriteCheck(
				&config.Config{Host: srv.URL, Token: "tok", Insecure: true},
				func(context.Context) error {
					checked++
					return denied
				},
			)
			req, err := http.NewRequestWithContext(context.Background(), method, srv.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := c.DoHTTP(req)

			safe := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
			if safe {
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()
				if checked != 0 || requests != 1 {
					t.Fatalf("checked=%d requests=%d, want 0 and 1", checked, requests)
				}
				return
			}

			if !errors.Is(err, denied) {
				t.Fatalf("error = %v, want denied", err)
			}
			if checked != 1 || requests != 0 {
				t.Fatalf("checked=%d requests=%d, want 1 and 0", checked, requests)
			}
		})
	}
}

func TestClient_WriteCheckCoversRequestAPIs(t *testing.T) {
	denied := errors.New("denied")
	c := client.NewWithWriteCheck(
		&config.Config{Host: "https://splunk.invalid", Token: "tok"},
		func(context.Context) error { return denied },
	)

	tests := map[string]func() error{
		"Do": func() error {
			_, err := c.Do(context.Background(), client.Request{Method: http.MethodPost, Path: "/test"})
			return err
		},
		"RawDo": func() error {
			_, _, err := c.RawDo(context.Background(), http.MethodPost, "/test", nil)
			return err
		},
		"RawDoAbs": func() error {
			_, _, err := c.RawDoAbs(context.Background(), http.MethodPost, "/test", nil)
			return err
		},
		"RawDoJSON": func() error {
			_, _, err := c.RawDoJSON(context.Background(), http.MethodPost, "/test", map[string]string{"key": "value"})
			return err
		},
		"Upload": func() error {
			_, err := c.Upload(context.Background(), "/test", "file", "test.txt", []byte("data"))
			return err
		},
	}

	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, denied) {
				t.Fatalf("error = %v, want denied", err)
			}
		})
	}
}

func TestClient_Token(t *testing.T) {
	cfg := &config.Config{Host: "https://splunk:8089", Token: "mytoken", Insecure: true}
	c := client.New(cfg)
	if c.Token() != "mytoken" {
		t.Errorf("expected token mytoken, got %q", c.Token())
	}
}

func TestClient_RawDoAbs_SendsFormBody(t *testing.T) {
	var gotContentType, gotBody string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	body := url.Values{}
	body.Set("name", "myindex")
	_, status, err := c.RawDoAbs(context.Background(), "POST", "/servicesNS/nobody/search/data/indexes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content-type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "name=myindex") {
		t.Errorf("expected name=myindex in body, got %q", gotBody)
	}
}

func TestClient_RawDo_SendsFormBody(t *testing.T) {
	var gotContentType, gotBody string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	body := url.Values{}
	body.Set("search", "index=main")
	_, status, err := c.RawDo(context.Background(), "POST", "/search/jobs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content-type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "search=") {
		t.Errorf("expected search param in body, got %q", gotBody)
	}
}

func TestClient_RawDoJSON_InvalidPayload(t *testing.T) {
	c := client.New(&config.Config{Host: "https://splunk:8089", Token: "tok", Insecure: true})
	_, _, err := c.RawDoJSON(context.Background(), "POST", "/path/", make(chan int))
	if err == nil {
		t.Fatal("expected error for unmarshalable payload")
	}
}

func TestClient_Upload_ReturnsErrorOn400(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"invalid file"}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Upload(context.Background(), "/data/lookup-table-files/", "file", "test.csv", []byte("a,b\n1,2"))
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestClient_Describe_ReturnsFields(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[{"fields":{"maxDataSizeMB":{"datatype":"Number"}}}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	fields, err := c.Describe(context.Background(), "/data/indexes/")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["maxDataSizeMB"]; !ok {
		t.Errorf("expected maxDataSizeMB in fields, got %v", fields)
	}
}

func TestClient_Describe_ReturnsErrorOn404(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"not found"}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Describe(context.Background(), "/data/indexes/")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClient_Describe_ReturnsErrorOnEmptyEntry(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Describe(context.Background(), "/data/indexes/")
	if err == nil {
		t.Fatal("expected error for empty entry response")
	}
}

func TestClient_RawDoAbs_SendsRequest(t *testing.T) {
	var gotQuery, gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("output_mode")
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, status, err := c.RawDoAbs(context.Background(), "GET", "/servicesNS/nobody/search/jobs/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if gotQuery != "json" {
		t.Errorf("expected output_mode=json, got %q", gotQuery)
	}
	if gotAuth != "Bearer tok" {
		t.Errorf("expected Bearer tok, got %q", gotAuth)
	}
}

func TestClient_UnsupportedMethodReturnsError(t *testing.T) {
	c := client.New(&config.Config{Host: "https://splunk:8089", Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "PATCH", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
	if !strings.Contains(err.Error(), "unsupported method") {
		t.Errorf("expected 'unsupported method' in error, got: %v", err)
	}
}

func TestClient_DeleteRequest(t *testing.T) {
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "DELETE", Path: "/data/indexes/", ID: "myindex"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("expected DELETE method, got %q", gotMethod)
	}
}

func TestClient_ParseEntries_InvalidJSON(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json at all`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestClient_ParseEntries_RawArray(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name":"main","maxDataSizeMB":500}]`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "main" {
		t.Errorf("expected name=main, got %v", results[0]["name"])
	}
}

func TestClient_PlainTextErrorBody(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "internal server error") {
		t.Errorf("expected plain text error in message, got: %v", err)
	}
}

func TestClient_401ReturnsAuthError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"unauthorized"}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "AUTH_ERROR") {
		t.Errorf("expected AUTH_ERROR in error, got: %v", err)
	}
}

func TestClient_429ReturnsRateLimited(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"rate limited"}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !strings.Contains(err.Error(), "RATE_LIMITED") {
		t.Errorf("expected RATE_LIMITED in error, got: %v", err)
	}
}

func TestClient_500ReturnsServerError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"messages":[{"type":"ERROR","text":"internal error"}]}`))
	}))
	defer srv.Close()

	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "SERVER_ERROR") {
		t.Errorf("expected SERVER_ERROR in error, got: %v", err)
	}
}

func TestClient_PostSendsFormBody(t *testing.T) {
	var gotContentType, gotBody string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entry":[]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Host: srv.URL, Token: "tok", Insecure: true}
	c := client.New(cfg)
	_, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/",
		Params: map[string]string{"name": "myindex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content-type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "name=myindex") {
		t.Errorf("expected name=myindex in body, got %q", gotBody)
	}
}
