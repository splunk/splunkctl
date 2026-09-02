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

func newWorkloadRuleTestServer(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
}

func runWorkloadRuleCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().StringArray("param", nil, "")
	parent := &cobra.Command{Use: "workload-rule"}
	parent.AddCommand(newWorkloadRuleAddCmd())
	root.AddCommand(parent)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestWorkloadRuleAdd_SendsNameNotRuleName(t *testing.T) {
	var gotName, gotRuleName string
	c := newWorkloadRuleTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/rules/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		gotName = r.FormValue("name")
		gotRuleName = r.FormValue("rule_name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "my_rule", "content": map[string]any{}}},
		})
	}))

	err := runWorkloadRuleCmd(t, c, []string{
		"workload-rule", "add", "my_rule", "--predicate", "role=admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "my_rule" {
		t.Errorf("expected name=my_rule, got %q", gotName)
	}
	if gotRuleName != "" {
		t.Errorf("expected no rule_name param, got %q", gotRuleName)
	}
}
