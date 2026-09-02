package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const fieldExtractionPath = "/data/props/extractions/"

// FieldExtractionCmd is the top-level field-extraction command. Registered in cmd/register_overrides.go.
var FieldExtractionCmd = &cobra.Command{
	Use:   "field-extraction",
	Short: "Manage field extractions",
}

var fieldExtractionListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List field extractions",
	Example:     `  splunkctl field-extraction list`,
	Annotations: map[string]string{"describe_path": fieldExtractionPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, fieldExtractionPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   fieldExtractionPath,
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

var fieldExtractionAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a field extraction",
	Example: `  splunkctl field-extraction add my_extraction --stanza access_combined --type EXTRACT --value "(?P<user>\w+) (?P<action>\w+)"
  splunkctl field-extraction add my_kv --stanza my_sourcetype --type REPORT --value my_transform`,
	Annotations: map[string]string{"describe_path": fieldExtractionPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, fieldExtractionPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		stanza, _ := cmd.Flags().GetString("stanza")
		extType, _ := cmd.Flags().GetString("type")
		value, _ := cmd.Flags().GetString("value")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":   name,
			"stanza": stanza,
			"type":   extType,
			"value":  value,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   fieldExtractionPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var fieldExtractionEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit a field extraction",
	Example:     `  splunkctl field-extraction edit my_extraction --value "(?P<user>\w+)"`,
	Annotations: map[string]string{"describe_path": fieldExtractionPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, fieldExtractionPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("value"); v != "" {
			params["value"] = v
		}
		if v, _ := cmd.Flags().GetString("stanza"); v != "" {
			params["stanza"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   fieldExtractionPath,
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

var fieldExtractionRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a field extraction",
	Example: `  splunkctl field-extraction remove my_extraction`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   fieldExtractionPath,
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
	for _, cmd := range []*cobra.Command{fieldExtractionListCmd, fieldExtractionAddCmd, fieldExtractionEditCmd, fieldExtractionRemoveCmd} {
		cmd.Flags().String("name", "", "extraction name (positional or --name)")
	}
	for _, cmd := range []*cobra.Command{fieldExtractionAddCmd, fieldExtractionEditCmd} {
		cmd.Flags().String("stanza", "", "sourcetype or source stanza (e.g. access_combined)")
		cmd.Flags().String("value", "", "regex pattern (EXTRACT) or transform name (REPORT)")
	}
	fieldExtractionAddCmd.Flags().String("type", "EXTRACT", "extraction type: EXTRACT or REPORT")
	FieldExtractionCmd.AddCommand(fieldExtractionListCmd, fieldExtractionAddCmd, fieldExtractionEditCmd, fieldExtractionRemoveCmd)
}
