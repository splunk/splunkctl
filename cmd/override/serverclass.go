package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const serverclassPath = "/deployment/server/serverclasses/"

// ServerclassCmd is the top-level serverclass command. Registered in cmd/register_overrides.go.
var ServerclassCmd = &cobra.Command{
	Use:   "serverclass",
	Short: "Manage deployment server serverclasses (Universal Forwarder fleet management)",
}

var serverclassListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List serverclasses",
	Example:     `  splunkctl serverclass list`,
	Annotations: map[string]string{"describe_path": serverclassPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, serverclassPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   serverclassPath,
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

var serverclassAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a serverclass",
	Example: `  splunkctl serverclass add linux_servers
  splunkctl serverclass add linux_servers --param whitelist.0=10.0.0.*`,
	Annotations: map[string]string{"describe_path": serverclassPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, serverclassPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{"name": name})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   serverclassPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var serverclassEditCmd = &cobra.Command{
	Use:         "edit <name>",
	Short:       "Edit a serverclass",
	Example:     `  splunkctl serverclass edit linux_servers --param whitelist.0=10.0.1.*`,
	Annotations: map[string]string{"describe_path": serverclassPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, serverclassPath); done {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   serverclassPath,
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

var serverclassRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a serverclass",
	Example: `  splunkctl serverclass remove linux_servers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   serverclassPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

var serverclassClientsCmd = &cobra.Command{
	Use:     "clients <name>",
	Short:   "List clients assigned to a serverclass",
	Example: `  splunkctl serverclass clients linux_servers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   serverclassPath + name + "/clients/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{serverclassListCmd, serverclassAddCmd, serverclassEditCmd, serverclassRemoveCmd, serverclassClientsCmd} {
		cmd.Flags().String("name", "", "serverclass name (positional or --name)")
	}
	ServerclassCmd.AddCommand(serverclassListCmd, serverclassAddCmd, serverclassEditCmd, serverclassRemoveCmd, serverclassClientsCmd)
}
