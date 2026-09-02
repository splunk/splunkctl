package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newClusterSearchHeadsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <guids>",
		Short: `Remove downed search heads from the cluster manager.`,
		Example: `
  'splunkctl cluster-search-heads remove-search-heads cluster-peers -guids GUID'
  'splunkctl cluster-search-heads remove-search-heads cluster-peers -guids GUID1,GUID2,GUID3'`,
		Annotations: map[string]string{"describe_path": "/cluster/master/control/default/remove_search_heads"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			impliedVal, _ := cmd.Flags().GetString("guids")
			if impliedVal == "" && len(args) > 0 {
				impliedVal = args[0]
			}
			if impliedVal != "" {
				params["guids"] = impliedVal
			}
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/cluster/master/control/default/remove_search_heads",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
	cmd.Flags().String("guids", "", "guids (positional or --guids)")
	return cmd
}

// ClusterSearchHeadsRemoveCmd overrides the generated remove to use POST
// instead of DELETE; the remove_search_heads custom action requires POST
// and rejects DELETE with "method expected=POST does not match actual=DELETE".
var ClusterSearchHeadsRemoveCmd = newClusterSearchHeadsRemoveCmd()
