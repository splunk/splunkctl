package override

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// RawCmd is the top-level raw command. Registered in cmd/register_overrides.go.
var RawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Make raw Splunk REST API calls (escape hatch)",
}

var rawGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "GET any /services/... path",
	Example: `  splunkctl raw get /server/info
  splunkctl raw get /data/indexes/ --param count=5
  splunkctl raw get /authentication/users/admin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		path := normalizeRawPath(args[0])
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   path,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var rawPostCmd = &cobra.Command{
	Use:   "post <path>",
	Short: "POST any /services/... path",
	Example: `  splunkctl raw post /data/indexes/ --param name=myindex
  splunkctl raw post /search/jobs --param search="search index=main"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		path := normalizeRawPath(args[0])
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   path,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var rawDeleteCmd = &cobra.Command{
	Use:     "delete <path>",
	Short:   "DELETE any /services/... path",
	Example: `  splunkctl raw delete /data/indexes/myindex`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		path := normalizeRawPath(args[0])
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   path,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), `{"deleted":true}`)
			return nil
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

// normalizeRawPath strips a leading /services prefix if the user included it,
// and ensures the path starts with /.
func normalizeRawPath(p string) string {
	p = strings.TrimPrefix(p, "/services")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func init() {
	RawCmd.AddCommand(rawGetCmd, rawPostCmd, rawDeleteCmd)
}
