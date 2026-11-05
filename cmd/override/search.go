package override

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// pollDeadlineGraceSec is added to the job's own max_time before we declare a
// client-side timeout, giving the server a chance to mark the job done after
// its own timer fires.
const pollDeadlineGraceSec = 30

// SearchCmd is the top-level search command. Registered in cmd/register_overrides.go.
var SearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Run a Splunk search and return results",
	Long: `Submit a Splunk search job and stream results to stdout.

Waits for the job to complete, then prints results.
Use --detach to return the job ID immediately.
Use --app to resolve app-local knowledge objects, with --owner for a user namespace.`,
	Example: `  splunkctl search 'index=main error'
  splunkctl search 'index=main' --maxout 500 --output json
  splunkctl search '| inputlookup my_lookup' --app my_app --owner nobody
  splunkctl search 'index=main sourcetype=access_combined' --detach`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

func init() {
	SearchCmd.Flags().Int("max-time", 0, "Max job runtime in seconds (0 = no limit)")
	SearchCmd.Flags().Int("maxout", 100, "Max results to return (max 50000)")
	SearchCmd.Flags().Bool("detach", false, "Submit job and return immediately with job ID")
	SearchCmd.Flags().String("app", "", "App namespace for resolving search knowledge objects")
	SearchCmd.Flags().String("owner", "nobody", "Namespace owner (used only with --app)")
	SearchCmd.Flags().String("workload-pool", "", "Workload pool name")
}

// searchExportCmd streams results directly without creating a persistent job.
var searchExportCmd = &cobra.Command{
	Use:   "export <query>",
	Short: "Stream search results without a persistent job (fastest for ad hoc queries)",
	Example: `  splunkctl search export 'index=main error earliest=-1h'
  splunkctl search export 'index=main' --maxout 1000 --output json
  splunkctl search export '| inputlookup my_lookup' --app my_app --output json`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		maxout, _ := cmd.Flags().GetInt("maxout")
		outFmt, _ := cmd.Flags().GetString("output")
		app, _ := cmd.Flags().GetString("app")
		owner, _ := cmd.Flags().GetString("owner")

		query := strings.Join(args, " ")
		if !strings.HasPrefix(strings.TrimSpace(query), "search ") &&
			!strings.HasPrefix(strings.TrimSpace(query), "|") {
			query = "search " + query
		}

		c := cmdctx.ClientFrom(cmd)
		form := url.Values{}
		form.Set("search", query)
		form.Set("output_mode", "json")
		if maxout > 0 {
			form.Set("count", fmt.Sprintf("%d", maxout))
		}

		body, status, err := c.RawDoAbs(cmd.Context(), "POST", searchJobsPath(app, owner)+"/export", form)
		if err != nil {
			return fmt.Errorf("export: %w", err)
		}
		if status >= 400 {
			return client.ErrorFromStatus(status, body)
		}

		// export returns newline-delimited JSON objects; collect results
		var results []map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(body)))
		for decoder.More() {
			var obj struct {
				Result map[string]any `json:"result"`
			}
			if err := decoder.Decode(&obj); err != nil {
				break
			}
			if obj.Result != nil {
				results = append(results, obj.Result)
			}
		}
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	searchExportCmd.Flags().Int("maxout", 100, "Max results to stream")
	searchExportCmd.Flags().String("app", "", "App namespace for resolving search knowledge objects")
	searchExportCmd.Flags().String("owner", "nobody", "Namespace owner (used only with --app)")
	SearchCmd.AddCommand(searchExportCmd)
}

// A form app= field does not select the knowledge-object namespace. Keep the
// selected endpoint for the entire job lifecycle, including control requests.
func searchJobsPath(app, owner string) string {
	if app == "" {
		return "/services/search/jobs"
	}
	return "/servicesNS/" + url.PathEscape(owner) + "/" + url.PathEscape(app) + "/search/jobs"
}

func runSearch(cmd *cobra.Command, args []string) error {
	maxTime, _ := cmd.Flags().GetInt("max-time")
	maxout, _ := cmd.Flags().GetInt("maxout")
	detach, _ := cmd.Flags().GetBool("detach")
	app, _ := cmd.Flags().GetString("app")
	owner, _ := cmd.Flags().GetString("owner")
	workload, _ := cmd.Flags().GetString("workload-pool")
	outFmt, _ := cmd.Flags().GetString("output")

	query := strings.Join(args, " ")
	if !strings.HasPrefix(strings.TrimSpace(query), "search ") &&
		!strings.HasPrefix(strings.TrimSpace(query), "|") {
		query = "search " + query
	}

	c := cmdctx.ClientFrom(cmd)
	jobsPath := searchJobsPath(app, owner)

	// Step 1: submit job
	form := url.Values{}
	form.Set("search", query)
	if maxTime > 0 {
		form.Set("max_time", fmt.Sprintf("%d", maxTime))
	}
	if workload != "" {
		form.Set("workload_pool", workload)
	}

	body, status, err := c.RawDoAbs(cmd.Context(), "POST", jobsPath, form)
	if err != nil {
		return fmt.Errorf("submitting search: %w", err)
	}
	if status >= 400 {
		return client.ErrorFromStatus(status, body)
	}

	var submitResp struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal(body, &submitResp); err != nil {
		return fmt.Errorf("parsing job submission response: %w", err)
	}
	sid := submitResp.SID

	if detach {
		// Extend TTL to 1 hour so the job survives a multi-turn conversation.
		c.RawDoAbs(cmd.Context(), "POST", jobsPath+"/"+url.PathEscape(sid)+"/control",
			url.Values{"action": {"setttl"}, "ttl": {"3600"}})
		fmt.Fprintf(cmd.OutOrStdout(), `{"job_id":%q,"ttl":3600}`+"\n", sid)
		return nil
	}

	// Step 2: poll until done
	if err := pollJob(cmd.Context(), c, jobsPath, sid, maxTime); err != nil {
		return err
	}

	// Step 3: fetch results
	return fetchResults(cmd, c, jobsPath, sid, maxout, outFmt)
}

func pollJob(ctx context.Context, c *client.Client, jobsPath, sid string, maxTimeSec int) error {
	var deadline time.Duration
	if maxTimeSec > 0 {
		deadline = time.Duration(maxTimeSec+pollDeadlineGraceSec) * time.Second
	}

	start := time.Now()
	for {
		if deadline > 0 && time.Since(start) > deadline {
			return fmt.Errorf("timed out waiting for job %s", sid)
		}

		// report progress to stderr only when it's a TTY (suppressed for agents)
		if term.IsTerminal(int(os.Stderr.Fd())) {
			fmt.Fprintf(os.Stderr, "\rwaiting for job %s... (elapsed: %s)", sid, time.Since(start).Round(time.Second))
		}

		body, status, err := c.RawDoAbs(ctx, "GET", jobsPath+"/"+url.PathEscape(sid), nil)
		if err != nil {
			return fmt.Errorf("polling job: %w", err)
		}
		if status >= 400 {
			return client.ErrorFromStatus(status, body)
		}

		var statusResp struct {
			Entry []struct {
				Content struct {
					DispatchState string `json:"dispatchState"`
					IsDone        bool   `json:"isDone"`
				} `json:"content"`
			} `json:"entry"`
		}
		if err := json.Unmarshal(body, &statusResp); err != nil {
			return fmt.Errorf("parsing job status: %w", err)
		}

		if len(statusResp.Entry) > 0 {
			content := statusResp.Entry[0].Content
			if content.IsDone {
				// check for failure before declaring success
				state := content.DispatchState
				if state == "FAILED" || state == "ERROR" {
					return fmt.Errorf("search job %s failed with state: %s", sid, state)
				}
				if term.IsTerminal(int(os.Stderr.Fd())) {
					fmt.Fprintln(os.Stderr) // clear the progress line
				}
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func fetchResults(cmd *cobra.Command, c *client.Client, jobsPath, sid string, maxout int, outFmt string) error {
	if maxout <= 0 || maxout > 50000 {
		maxout = 100
	}

	body, status, err := c.RawDoAbs(cmd.Context(), "GET",
		fmt.Sprintf("%s/%s/results?count=%d", jobsPath, url.PathEscape(sid), maxout), nil)
	if err != nil {
		return fmt.Errorf("fetching results: %w", err)
	}
	if status >= 400 {
		return client.ErrorFromStatus(status, body)
	}

	var resultsFeed struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(body, &resultsFeed); err != nil {
		return fmt.Errorf("parsing results: %w", err)
	}

	return output.Print(cmd.Context(), resultsFeed.Results, outFmt)
}
