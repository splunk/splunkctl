package override

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
	"github.com/splunk/splunkctl/internal/output"
)

func newAppTestServer(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	c := client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
	return c, srv
}

func runAppCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	app := &cobra.Command{Use: "app"}
	app.AddCommand(newInstallCmd())
	root.AddCommand(app)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestAppInstall_FilePathSendsFilenameTrue(t *testing.T) {
	var gotFilename string
	c, _ := newAppTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotFilename = r.FormValue("filename")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	if err := runAppCmd(t, c, []string{"app", "install", "/path/to/foo.tgz"}); err != nil {
		t.Fatal(err)
	}
	if gotFilename != "true" {
		t.Errorf("expected filename=true, got %q", gotFilename)
	}
}

func TestAppInstall_AppNameDoesNotSendFilenameTrue(t *testing.T) {
	var gotFilename string
	c, _ := newAppTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotFilename = r.FormValue("filename")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	if err := runAppCmd(t, c, []string{"app", "install", "myapp"}); err != nil {
		t.Fatal(err)
	}
	if gotFilename != "" {
		t.Errorf("expected no filename param, got %q", gotFilename)
	}
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/home/user/foo.tgz", true},
		{"./foo.tgz", true},
		{"foo.tgz", true},
		{"https://example.com/test_app.tgz?job=build&job_token=abc", false},
		{"foo.spl", true},
		{"foo.tar.gz", true},
		{"foo.tar", true},
		{"myapp", false},
		{"splunk-add-on-1.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isFilePath(tt.input); got != tt.expected {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
