package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const dataModelPath = "/data/models/"

// DataModelCmd is the top-level data-model command. Registered in cmd/register_overrides.go.
var DataModelCmd = &cobra.Command{
	Use:   "data-model",
	Short: "Manage data models",
}

var dataModelListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List data models",
	Example:     `  splunkctl data-model list`,
	Annotations: map[string]string{"describe_path": dataModelPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, dataModelPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   dataModelPath,
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

var dataModelAccelCmd = &cobra.Command{
	Use:   "acceleration",
	Short: "Show or set data model acceleration",
}

var dataModelAccelShowCmd = &cobra.Command{
	Use:     "show <name>",
	Short:   "Show acceleration status for a data model",
	Example: `  splunkctl data-model acceleration show Authentication`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   dataModelPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		// filter to acceleration fields
		for i, r := range results {
			filtered := map[string]any{"name": r["name"]}
			for _, k := range []string{"acceleration", "acceleration.earliest_time", "acceleration.backfill_time", "acceleration.max_time", "acceleration.schedule_priority"} {
				if v, ok := r[k]; ok {
					filtered[k] = v
				}
			}
			results[i] = filtered
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var dataModelAccelEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable acceleration for a data model",
	Example: `  splunkctl data-model acceleration enable Authentication
  splunkctl data-model acceleration enable Authentication --param acceleration.earliest_time=-1mon`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{"acceleration": "1"})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   dataModelPath,
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

var dataModelAccelDisableCmd = &cobra.Command{
	Use:     "disable <name>",
	Short:   "Disable acceleration for a data model",
	Example: `  splunkctl data-model acceleration disable Authentication`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   dataModelPath,
			ID:     name,
			Params: map[string]string{"acceleration": "0"},
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	dataModelListCmd.Flags().String("name", "", "data model name (positional or --name)")

	for _, cmd := range []*cobra.Command{dataModelAccelShowCmd, dataModelAccelEnableCmd, dataModelAccelDisableCmd} {
		cmd.Flags().String("name", "", "data model name (positional or --name)")
	}

	dataModelAccelCmd.AddCommand(dataModelAccelShowCmd, dataModelAccelEnableCmd, dataModelAccelDisableCmd)
	DataModelCmd.AddCommand(dataModelListCmd, dataModelAccelCmd)
}
