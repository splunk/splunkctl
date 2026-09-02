package override

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func urlValues(m map[string]string) url.Values {
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}

// JobsCmd is the top-level jobs command. Registered in cmd/register_overrides.go.
var JobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage Splunk search jobs",
}

var jobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent search jobs",
	Example: `  splunkctl jobs list
  splunkctl jobs list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/search/jobs",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var jobsShowCmd = &cobra.Command{
	Use:   "show <sid>",
	Short: "Show status and results of a search job",
	Long:  "Poll until the job is done, then fetch and display results.",
	Example: `  splunkctl jobs show rt_search_1234
  splunkctl jobs show rt_search_1234 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sid := args[0]
		c := cmdctx.ClientFrom(cmd)
		outFmt, _ := cmd.Flags().GetString("output")

		// Check the job exists before polling
		body, status, err := c.RawDo(cmd.Context(), "GET", "/search/jobs/"+sid, nil)
		if err != nil {
			return err
		}
		if status == 404 {
			return fmt.Errorf("job %s not found — it may have expired (TTL exceeded)", sid)
		}
		if status >= 400 {
			return fmt.Errorf("job status check failed (%d): %s", status, string(body))
		}

		// poll until done
		if err := pollJob(cmd.Context(), c, sid, 0); err != nil {
			return err
		}
		// fetch results with default maxout
		return fetchResults(cmd, c, sid, 100, outFmt)
	},
}

var jobsCancelCmd = &cobra.Command{
	Use:     "cancel <sid>",
	Short:   "Cancel a running search job",
	Example: `  splunkctl jobs cancel rt_search_1234`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sid := args[0]
		c := cmdctx.ClientFrom(cmd)

		_, status, err := c.RawDo(cmd.Context(), "DELETE", "/search/jobs/"+sid, nil)
		if err != nil {
			return fmt.Errorf("cancelling job: %w", err)
		}
		if status >= 400 {
			return fmt.Errorf("cancel failed (status %d)", status)
		}
		fmt.Fprintf(cmd.OutOrStdout(), `{"job_id":%q,"cancelled":true}`+"\n", sid)
		return nil
	},
}

var jobsStatusCmd = &cobra.Command{
	Use:     "status <sid>",
	Short:   "Show current state of a search job without fetching results",
	Example: `  splunkctl jobs status rt_search_1234`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sid := args[0]
		c := cmdctx.ClientFrom(cmd)

		body, status, err := c.RawDo(cmd.Context(), "GET", "/search/jobs/"+sid, nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("status check failed (%d): %s", status, string(body))
		}

		var resp struct {
			Entry []struct {
				Content map[string]any `json:"content"`
			} `json:"entry"`
		}
		if err := unmarshalJSON(body, &resp); err != nil {
			return err
		}
		if len(resp.Entry) == 0 {
			return fmt.Errorf("job %s not found", sid)
		}
		content := resp.Entry[0].Content
		summary := map[string]any{
			"sid":           sid,
			"dispatchState": content["dispatchState"],
			"isDone":        content["isDone"],
			"eventCount":    content["eventCount"],
			"resultCount":   content["resultCount"],
			"runDuration":   content["runDuration"],
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{summary}, outFmt)
	},
}

var jobsEventsCmd = &cobra.Command{
	Use:   "events <sid>",
	Short: "Fetch raw events (pre-transform) from a completed search job",
	Example: `  splunkctl jobs events rt_search_1234
  splunkctl jobs events rt_search_1234 --maxout 500`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sid := args[0]
		maxout, _ := cmd.Flags().GetInt("maxout")
		if maxout <= 0 || maxout > 50000 {
			maxout = 100
		}
		c := cmdctx.ClientFrom(cmd)

		body, status, err := c.RawDo(cmd.Context(), "GET",
			fmt.Sprintf("/search/jobs/%s/events?count=%d&output_mode=json", sid, maxout), nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("events fetch failed (%d): %s", status, string(body))
		}
		var feed struct {
			Results []map[string]any `json:"results"`
		}
		if err := unmarshalJSON(body, &feed); err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), feed.Results, outFmt)
	},
}

var jobsPauseCmd = &cobra.Command{
	Use:   "pause <sid>",
	Short: "Pause a running search job",
	Args:  cobra.ExactArgs(1),
	RunE:  jobControl("pause"),
}

var jobsUnpauseCmd = &cobra.Command{
	Use:   "unpause <sid>",
	Short: "Resume a paused search job",
	Args:  cobra.ExactArgs(1),
	RunE:  jobControl("unpause"),
}

var jobsFinalizeCmd = &cobra.Command{
	Use:   "finalize <sid>",
	Short: "Finalize a search job (stop and make results available)",
	Args:  cobra.ExactArgs(1),
	RunE:  jobControl("finalize"),
}

func jobControl(action string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sid := args[0]
		c := cmdctx.ClientFrom(cmd)
		form := map[string]string{"action": action}
		_, status, err := c.RawDo(cmd.Context(), "POST", "/search/jobs/"+sid+"/control",
			urlValues(form))
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("%s failed (status %d)", action, status)
		}
		fmt.Fprintf(cmd.OutOrStdout(), `{"sid":%q,"action":%q}`+"\n", sid, action)
		return nil
	}
}

func init() {
	jobsEventsCmd.Flags().Int("maxout", 100, "Max events to return")
	JobsCmd.AddCommand(jobsListCmd, jobsShowCmd, jobsCancelCmd, jobsStatusCmd, jobsEventsCmd, jobsPauseCmd, jobsUnpauseCmd, jobsFinalizeCmd)
}
