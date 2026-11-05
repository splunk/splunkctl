package override_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/splunk/splunkctl/cmd/override"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func TestSearch_SubmitJob(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/search/jobs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("search") == "" {
			t.Error("search param required")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "1234567890.123", "content": map[string]any{"sid": "1234567890.123"}},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "POST", "/search/jobs", url.Values{
		"search":      []string{"search index=main error"},
		"output_mode": []string{"json"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}

func TestSearch_PollJob(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/search/jobs/1234567890.123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "1234567890.123",
					"content": map[string]any{
						"sid":           "1234567890.123",
						"isDone":        true,
						"dispatchState": "DONE",
						"resultCount":   3,
					},
				},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/search/jobs/1234567890.123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	entries := feed["entry"].([]any)
	content := entries[0].(map[string]any)["content"].(map[string]any)
	if content["isDone"] != true {
		t.Errorf("expected isDone=true")
	}
}

func TestSearch_FetchResults(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/search/jobs/1234567890.123/results" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"results": []any{
				map[string]any{"host": "myhost", "source": "/var/log/app.log", "_raw": "error occurred"},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/search/jobs/1234567890.123/results", url.Values{
		"count": []string{"100"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	results := feed["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

// Use the production handlers and flag definitions on fresh command adapters.
// The existing global flags share their values, so restore them after each test
// and keep these command tests sequential.
func runSearchNamespaceCommand(t *testing.T, c *client.Client, args []string) (string, error) {
	t.Helper()
	copyCommand := func(source *cobra.Command) *cobra.Command {
		t.Helper()
		cmd := &cobra.Command{Use: source.Use, Args: source.Args, RunE: source.RunE}
		source.Flags().VisitAll(func(flag *pflag.Flag) {
			value, changed := flag.Value.String(), flag.Changed
			t.Cleanup(func() {
				if err := flag.Value.Set(value); err != nil {
					t.Errorf("restore --%s: %v", flag.Name, err)
				}
				flag.Changed = changed
			})
			if err := flag.Value.Set(flag.DefValue); err != nil {
				t.Fatalf("reset --%s: %v", flag.Name, err)
			}
			flag.Changed = false
			cmd.Flags().AddFlag(flag)
		})
		return cmd
	}

	var out bytes.Buffer
	root := &cobra.Command{Use: "splunkctl", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("output", "", "")
	search := copyCommand(override.SearchCmd)
	for _, child := range override.SearchCmd.Commands() {
		search.AddCommand(copyCommand(child))
	}
	root.AddCommand(search)
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetContext(output.WithWriter(cmdctx.WithClient(context.Background(), c), &out))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestSearchNamespaceLifecycle(t *testing.T) {
	const query = "| inputlookup my_lookup"
	for _, scope := range []struct {
		name, base string
		flags      []string
	}{
		{"default", "/services/search/jobs", nil},
		{"owner without app", "/services/search/jobs", []string{"--owner", "alice"}},
		{"app", "/servicesNS/nobody/my_app/search/jobs", []string{"--app", "my_app"}},
		{"private", "/servicesNS/alice/my_app/search/jobs", []string{"--app", "my_app", "--owner", "alice"}},
		{"escaped segments", "/servicesNS/alice%2Fteam/my%20app%2F%23%3F/search/jobs", []string{"--app", "my app/#?", "--owner", "alice/team"}},
	} {
		for _, mode := range []string{"sync", "export", "detach"} {
			t.Run(scope.name+"/"+mode, func(t *testing.T) {
				requests := make(chan string, 8)
				c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests <- r.Method + " " + r.URL.EscapedPath()
					if r.URL.Query().Get("output_mode") != "json" {
						t.Error("request must use output_mode=json")
					}
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					if r.PostForm.Has("app") || r.PostForm.Has("owner") {
						t.Error("app and owner belong in the namespace path, not the form")
					}
					w.Header().Set("Content-Type", "application/json")
					switch r.Method + " " + r.URL.EscapedPath() {
					case "POST " + scope.base, "POST " + scope.base + "/export":
						if r.PostForm.Get("search") != query {
							t.Errorf("search = %q, want %q", r.PostForm.Get("search"), query)
						}
						if mode == "export" {
							fmt.Fprintln(w, `{"result":{"marker":"app-local"}}`)
						} else {
							fmt.Fprint(w, `{"sid":"test.sid"}`)
						}
					case "GET " + scope.base + "/test.sid":
						fmt.Fprint(w, `{"entry":[{"content":{"isDone":true,"dispatchState":"DONE"}}]}`)
					case "GET " + scope.base + "/test.sid/results":
						fmt.Fprint(w, `{"results":[{"marker":"app-local"}]}`)
					case "POST " + scope.base + "/test.sid/control":
						if r.PostForm.Get("action") != "setttl" || r.PostForm.Get("ttl") != "3600" {
							t.Errorf("unexpected control form: %v", r.PostForm)
						}
						fmt.Fprint(w, `{}`)
					default:
						http.Error(w, "unexpected namespace", http.StatusNotFound)
					}
				}))

				args := []string{"search"}
				if mode == "export" {
					args = append(args, "export")
				}
				args = append(args, query, "--output", "json")
				args = append(args, scope.flags...)
				if mode == "detach" {
					args = append(args, "--detach")
				}
				out, err := runSearchNamespaceCommand(t, c, args)
				if err != nil {
					t.Fatalf("search failed: %v; output: %s", err, out)
				}
				wantRequests := []string{"POST " + scope.base, "GET " + scope.base + "/test.sid", "GET " + scope.base + "/test.sid/results"}
				if mode == "export" {
					wantRequests = []string{"POST " + scope.base + "/export"}
				} else if mode == "detach" {
					wantRequests = []string{"POST " + scope.base, "POST " + scope.base + "/test.sid/control"}
				}
				var gotRequests []string
				for len(requests) > 0 {
					gotRequests = append(gotRequests, <-requests)
				}
				if !reflect.DeepEqual(gotRequests, wantRequests) {
					t.Errorf("requests = %v, want %v", gotRequests, wantRequests)
				}
				if mode == "detach" {
					var got struct {
						JobID string `json:"job_id"`
						TTL   int    `json:"ttl"`
					}
					if err := json.Unmarshal([]byte(out), &got); err != nil || got.JobID != "test.sid" || got.TTL != 3600 {
						t.Fatalf("unexpected detach output %q: %v", out, err)
					}
				} else {
					var got []map[string]string
					if err := json.Unmarshal([]byte(out), &got); err != nil || !reflect.DeepEqual(got, []map[string]string{{"marker": "app-local"}}) {
						t.Fatalf("unexpected search output %q: %v", out, err)
					}
				}
			})
		}
	}
}
