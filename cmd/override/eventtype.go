package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const eventtypePath = "/saved/eventtypes/"

// EventtypeCmd is the top-level eventtype command. Registered in cmd/register_overrides.go.
var EventtypeCmd = &cobra.Command{
	Use:   "eventtype",
	Short: "Manage event types",
}

var eventtypeListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List event types",
	Example:     `  splunkctl eventtype list`,
	Annotations: map[string]string{"describe_path": eventtypePath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, eventtypePath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   eventtypePath,
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

var eventtypeAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create an event type",
	Example: `  splunkctl eventtype add web_errors --search 'index=main sourcetype=access_combined status>=400'
  splunkctl eventtype add web_errors --search 'index=main status>=400' --param priority=1`,
	Annotations: map[string]string{"describe_path": eventtypePath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, eventtypePath); done {
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
			Path:   eventtypePath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var eventtypeEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit an event type",
	Example:     `  splunkctl eventtype edit web_errors --search 'index=main status>=500'`,
	Annotations: map[string]string{"describe_path": eventtypePath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, eventtypePath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("search"); v != "" {
			params["search"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   eventtypePath,
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

var eventtypeRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete an event type",
	Example: `  splunkctl eventtype remove web_errors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   eventtypePath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{eventtypeListCmd, eventtypeAddCmd, eventtypeEditCmd, eventtypeRemoveCmd} {
		cmd.Flags().String("name", "", "event type name (positional or --name)")
	}
	eventtypeAddCmd.Flags().String("search", "", "SPL search string (required)")
	eventtypeEditCmd.Flags().String("search", "", "new SPL search string")
	EventtypeCmd.AddCommand(eventtypeListCmd, eventtypeAddCmd, eventtypeEditCmd, eventtypeRemoveCmd)
}
