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

func newWorkloadPoolTestServer(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
}

func runWorkloadPoolCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().StringArray("param", nil, "")
	parent := &cobra.Command{Use: "workload-pool"}
	parent.AddCommand(newWorkloadPoolAddCmd())
	root.AddCommand(parent)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestWorkloadPoolAdd_SendsNameNotPoolName(t *testing.T) {
	var gotName, gotPoolName string
	c := newWorkloadPoolTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/pools/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		gotName = r.FormValue("name")
		gotPoolName = r.FormValue("pool_name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "pool_a", "content": map[string]any{}}},
		})
	}))

	err := runWorkloadPoolCmd(t, c, []string{
		"workload-pool", "add", "pool_a",
		"--cpu_weight", "30", "--mem_weight", "20", "--category", "search",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "pool_a" {
		t.Errorf("expected name=pool_a, got %q", gotName)
	}
	if gotPoolName != "" {
		t.Errorf("expected no pool_name param, got %q", gotPoolName)
	}
}
