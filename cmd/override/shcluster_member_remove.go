package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newShclusterMemberRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: `Remove the specified member if run on the captain or, if run on a non-captain member, remove that member from the search head cluster.`,
		Example: `
  splunkctl shcluster-member remove -mgmt_uri https://myserver:1234`,
		Annotations: map[string]string{"describe_path": "/shcluster/member/consensus/default/remove_server"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			if v, _ := cmd.Flags().GetString("mgmt_uri"); v != "" {
				params["mgmt_uri"] = v
			}
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/shcluster/member/consensus/default/remove_server",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
	cmd.Flags().String("mgmt_uri", "", `If run on the captain, this is the management uri of the member to be removed.`)
	return cmd
}

// ShclusterMemberRemoveCmd overrides the generated remove to use POST instead
// of DELETE; the /shcluster/member/consensus/default/remove_server custom
// action requires POST and rejects DELETE with "requires POST".
var ShclusterMemberRemoveCmd = newShclusterMemberRemoveCmd()
