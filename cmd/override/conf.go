package override

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// ConfCmd is the top-level conf command. Registered in cmd/register_overrides.go.
var ConfCmd = &cobra.Command{
	Use:   "conf",
	Short: "Read and write Splunk .conf file stanzas via REST",
}

var confGetCmd = &cobra.Command{
	Use:   "get <file> [stanza]",
	Short: "Read a conf file or specific stanza",
	Example: `  splunkctl conf get inputs
  splunkctl conf get inputs monitor://var/log
  splunkctl conf get transforms lookup_name`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		file := args[0]
		stanza := ""
		if len(args) > 1 {
			stanza = args[1]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/configs/conf-" + file + "/",
			ID:     stanza,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var confSetCmd = &cobra.Command{
	Use:   "set <file> <stanza> --param key=value",
	Short: "Set key=value pairs in a conf file stanza",
	Example: `  splunkctl conf set inputs "monitor:///var/log/app.log" --param index=main --param sourcetype=app_log
  splunkctl conf set limits search --param max_searches_per_cpu=4`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		file, stanza := args[0], args[1]
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if len(params) == 0 {
			return fmt.Errorf("at least one --param key=value is required")
		}
		// Try creating first (POST to base path with name=stanza).
		// If stanza already exists (409), fall back to updating (POST to named path).
		params["name"] = stanza
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   "/configs/conf-" + file + "/",
			Params: params,
		})
		if err != nil {
			var splunkErr *client.SplunkError
			if errors.As(err, &splunkErr) && splunkErr.Status == 409 {
				delete(params, "name")
				results, err = c.Do(cmd.Context(), client.Request{
					Method: "POST",
					Path:   "/configs/conf-" + file + "/",
					ID:     stanza,
					Params: params,
				})
			}
		}
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var confDeleteCmd = &cobra.Command{
	Use:     "delete <file> <stanza>",
	Short:   "Delete a stanza from a conf file",
	Example: `  splunkctl conf delete inputs "monitor:///var/log/old.log"`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		file, stanza := args[0], args[1]
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   "/configs/conf-" + file + "/",
			ID:     stanza,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"file": file, "stanza": stanza, "deleted": true}}, outFmt)
	},
}

func init() {
	ConfCmd.AddCommand(confGetCmd, confSetCmd, confDeleteCmd)
}
