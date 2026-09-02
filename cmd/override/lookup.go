package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const lookupPath = "/data/transforms/lookups/"

// LookupCmd is the top-level lookup command. Registered in cmd/register_overrides.go.
var LookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Manage lookup transform definitions",
}

var lookupListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List lookup definitions",
	Example:     `  splunkctl lookup list`,
	Annotations: map[string]string{"describe_path": lookupPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, lookupPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   lookupPath,
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

var lookupAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a lookup definition",
	Example: `  splunkctl lookup add my_lookup --filename my_lookup.csv
  splunkctl lookup add my_lookup --filename my_lookup.csv --param case_sensitive_match=true`,
	Annotations: map[string]string{"describe_path": lookupPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, lookupPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		filename, _ := cmd.Flags().GetString("filename")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":     name,
			"filename": filename,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   lookupPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var lookupEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit a lookup definition",
	Example:     `  splunkctl lookup edit my_lookup --filename new_file.csv`,
	Annotations: map[string]string{"describe_path": lookupPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, lookupPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("filename"); v != "" {
			params["filename"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   lookupPath,
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

var lookupRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a lookup definition",
	Example: `  splunkctl lookup remove my_lookup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   lookupPath,
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
	for _, cmd := range []*cobra.Command{lookupListCmd, lookupAddCmd, lookupEditCmd, lookupRemoveCmd} {
		cmd.Flags().String("name", "", "lookup name (positional or --name)")
	}
	for _, cmd := range []*cobra.Command{lookupAddCmd, lookupEditCmd} {
		cmd.Flags().String("filename", "", "lookup table filename (e.g. my_lookup.csv)")
	}
	LookupCmd.AddCommand(lookupListCmd, lookupAddCmd, lookupEditCmd, lookupRemoveCmd)
}
