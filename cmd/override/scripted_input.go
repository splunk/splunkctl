package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const scriptedInputPath = "/data/inputs/script/"

// ScriptedInputCmd is the top-level scripted-input command. Registered in cmd/register_overrides.go.
var ScriptedInputCmd = &cobra.Command{
	Use:   "scripted-input",
	Short: "Manage scripted data inputs",
}

var scriptedInputListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List scripted inputs",
	Example:     `  splunkctl scripted-input list`,
	Annotations: map[string]string{"describe_path": scriptedInputPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, scriptedInputPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   scriptedInputPath,
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

var scriptedInputAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a scripted input",
	Example: `  splunkctl scripted-input add /opt/splunk/etc/apps/myapp/bin/collect.py --interval 60
  splunkctl scripted-input add myscript.sh --interval 300 --param index=main --param sourcetype=my_script`,
	Annotations: map[string]string{"describe_path": scriptedInputPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, scriptedInputPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		interval, _ := cmd.Flags().GetString("interval")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":     name,
			"interval": interval,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   scriptedInputPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var scriptedInputEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit a scripted input",
	Example:     `  splunkctl scripted-input edit myscript.sh --interval 120`,
	Annotations: map[string]string{"describe_path": scriptedInputPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, scriptedInputPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		if v, _ := cmd.Flags().GetString("interval"); v != "" {
			params["interval"] = v
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   scriptedInputPath,
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

var scriptedInputRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a scripted input",
	Example: `  splunkctl scripted-input remove myscript.sh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   scriptedInputPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

var scriptedInputEnableCmd = &cobra.Command{
	Use:     "enable <name>",
	Short:   "Enable a scripted input",
	Example: `  splunkctl scripted-input enable myscript.sh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   scriptedInputPath + name + "/enable",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var scriptedInputDisableCmd = &cobra.Command{
	Use:     "disable <name>",
	Short:   "Disable a scripted input",
	Example: `  splunkctl scripted-input disable myscript.sh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   scriptedInputPath + name + "/disable",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{scriptedInputListCmd, scriptedInputAddCmd, scriptedInputEditCmd, scriptedInputRemoveCmd, scriptedInputEnableCmd, scriptedInputDisableCmd} {
		cmd.Flags().String("name", "", "script path or name (positional or --name)")
	}
	scriptedInputAddCmd.Flags().String("interval", "60", "execution interval in seconds")
	scriptedInputEditCmd.Flags().String("interval", "", "new execution interval in seconds")
	ScriptedInputCmd.AddCommand(scriptedInputListCmd, scriptedInputAddCmd, scriptedInputEditCmd, scriptedInputRemoveCmd, scriptedInputEnableCmd, scriptedInputDisableCmd)
}
