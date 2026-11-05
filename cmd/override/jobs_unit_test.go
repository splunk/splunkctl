package override

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
	"github.com/splunk/splunkctl/internal/output"
)

func TestJobsList_ForwardsCountParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("count"); got != "5" {
			t.Errorf("count query parameter = %q, want 5", got)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"entry":[]}`)
	}))
	t.Cleanup(srv.Close)
	c := client.New(&config.Config{Host: srv.URL, Token: "test-token"})
	cmd := &cobra.Command{Use: "list", RunE: jobsListCmd.RunE}
	cmd.Flags().StringArray("param", nil, "")
	cmd.Flags().String("output", "json", "")
	cmd.SetContext(output.WithWriter(cmdctx.WithClient(context.Background(), c), io.Discard))
	cmd.SetArgs([]string{"--param", "count=5"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalJSON_ValidJSON(t *testing.T) {
	var m map[string]string
	err := unmarshalJSON([]byte(`{"key":"value"}`), &m)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("expected key=value, got %q", m["key"])
	}
}

func TestUnmarshalJSON_InvalidJSON(t *testing.T) {
	var m map[string]string
	err := unmarshalJSON([]byte(`not json`), &m)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	var syntaxErr *json.SyntaxError
	if err == syntaxErr {
		t.Error("expected a JSON syntax error")
	}
}

func TestUrlValues_ConvertsMapsCorrectly(t *testing.T) {
	v := urlValues(map[string]string{"search": "index=main", "count": "10"})
	if v.Get("search") != "index=main" {
		t.Errorf("expected search=index=main, got %q", v.Get("search"))
	}
	if v.Get("count") != "10" {
		t.Errorf("expected count=10, got %q", v.Get("count"))
	}
}
