package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const distSearchPeerPath = "/search/distributed/peers/"
const distSearchConfigPath = "/search/distributed/config/"

// DistSearchCmd is the top-level dist-search command. Registered in cmd/register_overrides.go.
var DistSearchCmd = &cobra.Command{
	Use:   "dist-search",
	Short: "Manage distributed search peers",
}

var distSearchPeerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List distributed search peers",
	Example: `  splunkctl dist-search peer list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   distSearchPeerPath,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var distSearchPeerAddCmd = &cobra.Command{
	Use:   "add <host:port>",
	Short: "Add a search peer",
	Example: `  splunkctl dist-search peer add indexer1.example.com:9997
  splunkctl dist-search peer add indexer1.example.com:9997 --param remotePassword=changeme`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		peer, _ := cmd.Flags().GetString("peer")
		if peer == "" && len(args) > 0 {
			peer = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{"name": peer})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   distSearchPeerPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var distSearchPeerRemoveCmd = &cobra.Command{
	Use:     "remove <host:port>",
	Short:   "Remove a search peer",
	Example: `  splunkctl dist-search peer remove indexer1.example.com:9997`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		peer, _ := cmd.Flags().GetString("peer")
		if peer == "" && len(args) > 0 {
			peer = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   distSearchPeerPath,
			ID:     peer,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"peer": peer, "deleted": true}}, outFmt)
	},
}

var distSearchConfigCmd = &cobra.Command{
	Use:     "config",
	Short:   "Show distributed search configuration",
	Example: `  splunkctl dist-search config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   distSearchConfigPath,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var distSearchPeerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Manage distributed search peers",
}

func init() {
	distSearchPeerAddCmd.Flags().String("peer", "", "peer host:port (positional or --peer)")
	distSearchPeerRemoveCmd.Flags().String("peer", "", "peer host:port (positional or --peer)")
	distSearchPeerCmd.AddCommand(distSearchPeerListCmd, distSearchPeerAddCmd, distSearchPeerRemoveCmd)
	DistSearchCmd.AddCommand(distSearchPeerCmd, distSearchConfigCmd)
}
