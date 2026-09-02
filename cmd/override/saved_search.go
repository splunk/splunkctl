package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const savedSearchPath = "/saved/searches/"

// SavedSearchCmd is the top-level saved-search command. Registered in cmd/register_overrides.go.
var SavedSearchCmd = &cobra.Command{
	Use:   "saved-search",
	Short: "Manage saved searches and alerts",
}

var savedSearchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved searches",
	Example: `  splunkctl saved-search list
  splunkctl saved-search list --output json`,
	Annotations: map[string]string{"describe_path": savedSearchPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, savedSearchPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   savedSearchPath,
			ID:     name,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var savedSearchAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a saved search",
	Example: `  splunkctl saved-search add "My Search" --search 'index=main error'
  splunkctl saved-search add "Hourly Errors" --search 'index=main error | stats count' --param cron_schedule="0 * * * *" --param is_scheduled=1`,
	Annotations: map[string]string{"describe_path": savedSearchPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, savedSearchPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		search, _ := cmd.Flags().GetString("search")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":   name,
			"search": search,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   savedSearchPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var savedSearchEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a saved search",
	Example: `  splunkctl saved-search edit "My Search" --search 'index=main error | head 100'
  splunkctl saved-search edit "My Search" --param is_scheduled=1 --param cron_schedule="0 * * * *"`,
	Annotations: map[string]string{"describe_path": savedSearchPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, savedSearchPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if s, _ := cmd.Flags().GetString("search"); s != "" {
			params["search"] = s
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   savedSearchPath,
			ID:     name,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var savedSearchRemoveCmd = &cobra.Command{
	Use:         "remove <name>",
	Short:       "Delete a saved search",
	Example:     `  splunkctl saved-search remove "My Search"`,
	Annotations: map[string]string{"describe_path": savedSearchPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   savedSearchPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

var savedSearchDispatchCmd = &cobra.Command{
	Use:   "dispatch <name>",
	Short: "Run a saved search manually and return the job ID",
	Example: `  splunkctl saved-search dispatch "Errors in the last hour"
  splunkctl saved-search dispatch "My Report" --param dispatch.earliest_time=-24h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   savedSearchPath + name + "/dispatch",
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{savedSearchListCmd, savedSearchAddCmd, savedSearchEditCmd, savedSearchDispatchCmd} {
		cmd.Flags().String("name", "", "saved search name (positional or --name)")
	}
	savedSearchRemoveCmd.Flags().String("name", "", "saved search name (positional or --name)")
	savedSearchAddCmd.Flags().String("search", "", "SPL search string (required)")
	savedSearchEditCmd.Flags().String("search", "", "new SPL search string")

	SavedSearchCmd.AddCommand(savedSearchListCmd, savedSearchAddCmd, savedSearchEditCmd, savedSearchRemoveCmd, savedSearchDispatchCmd)
}
