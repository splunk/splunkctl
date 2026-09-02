package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const hecPath = "/data/inputs/http/"

// HecCmd is the top-level hec command. Registered in cmd/register_overrides.go.
var HecCmd = &cobra.Command{
	Use:   "hec",
	Short: "Manage HTTP Event Collector (HEC) tokens",
}

var hecListCmd = &cobra.Command{
	Use:   "list",
	Short: "List HEC tokens",
	Example: `  splunkctl hec list
  splunkctl hec list --output json`,
	Annotations: map[string]string{"describe_path": hecPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, hecPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   hecPath,
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

var hecAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a HEC token",
	Example: `  splunkctl hec add mytoken
  splunkctl hec add mytoken --index main --sourcetype json
  splunkctl hec add mytoken --index main --param useACK=1`,
	Annotations: map[string]string{"describe_path": hecPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, hecPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{"name": name})
		if v, _ := cmd.Flags().GetString("index"); v != "" {
			params["index"] = v
		}
		if v, _ := cmd.Flags().GetString("sourcetype"); v != "" {
			params["sourcetype"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   hecPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var hecEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a HEC token",
	Example: `  splunkctl hec edit mytoken --index newindex
  splunkctl hec edit mytoken --param useACK=1`,
	Annotations: map[string]string{"describe_path": hecPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, hecPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("index"); v != "" {
			params["index"] = v
		}
		if v, _ := cmd.Flags().GetString("sourcetype"); v != "" {
			params["sourcetype"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   hecPath,
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

var hecRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a HEC token",
	Example: `  splunkctl hec remove mytoken`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   hecPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

var hecEnableCmd = &cobra.Command{
	Use:     "enable <name>",
	Short:   "Enable a HEC token",
	Example: `  splunkctl hec enable mytoken`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   hecPath + name + "/enable",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var hecDisableCmd = &cobra.Command{
	Use:     "disable <name>",
	Short:   "Disable a HEC token",
	Example: `  splunkctl hec disable mytoken`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   hecPath + name + "/disable",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{hecListCmd, hecAddCmd, hecEditCmd, hecRemoveCmd, hecEnableCmd, hecDisableCmd} {
		cmd.Flags().String("name", "", "HEC token name (positional or --name)")
	}
	for _, cmd := range []*cobra.Command{hecAddCmd, hecEditCmd} {
		cmd.Flags().String("index", "", "default index for events")
		cmd.Flags().String("sourcetype", "", "default sourcetype for events")
	}

	HecCmd.AddCommand(hecListCmd, hecAddCmd, hecEditCmd, hecRemoveCmd, hecEnableCmd, hecDisableCmd)
}
