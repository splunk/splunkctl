package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const macroPath = "/admin/macros/"

// MacroCmd is the top-level macro command. Registered in cmd/register_overrides.go.
var MacroCmd = &cobra.Command{
	Use:   "macro",
	Short: "Manage search macros",
}

var macroListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List search macros",
	Example:     `  splunkctl macro list`,
	Annotations: map[string]string{"describe_path": macroPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, macroPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   macroPath,
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

var macroAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a search macro",
	Example: `  splunkctl macro add my_macro --definition 'index=main sourcetype=syslog'
  splunkctl macro add my_macro(1) --definition 'index=main $arg1$' --param args=arg1`,
	Annotations: map[string]string{"describe_path": macroPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, macroPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		definition, _ := cmd.Flags().GetString("definition")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":       name,
			"definition": definition,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   macroPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var macroEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit a search macro",
	Example:     `  splunkctl macro edit my_macro --definition 'index=main sourcetype=access_combined'`,
	Annotations: map[string]string{"describe_path": macroPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, macroPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("definition"); v != "" {
			params["definition"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   macroPath,
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

var macroRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a search macro",
	Example: `  splunkctl macro remove my_macro`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   macroPath,
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
	for _, cmd := range []*cobra.Command{macroListCmd, macroAddCmd, macroEditCmd, macroRemoveCmd} {
		cmd.Flags().String("name", "", "macro name (positional or --name)")
	}
	macroAddCmd.Flags().String("definition", "", "SPL definition for the macro (required)")
	macroEditCmd.Flags().String("definition", "", "new SPL definition")
	MacroCmd.AddCommand(macroListCmd, macroAddCmd, macroEditCmd, macroRemoveCmd)
}
