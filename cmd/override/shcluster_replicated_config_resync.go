package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newShclusterReplicatedConfigResyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resync",
		Short: `Destructively resyncs this node to the latest replicated config on the captain.`,
		Example: `
  'splunkctl shcluster-replicated-config resync`,
		Annotations: map[string]string{"describe_path": "/replication/configuration/commits"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			params["resync_destructive"] = "true"
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/replication/configuration/commits",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
}

// ShclusterReplicatedConfigResyncCmd overrides the generated resync to send
// "resync_destructive=true"; the /replication/configuration/commits endpoint
// requires either "push_to" or "resync_destructive" and rejects requests
// without one ("Missing required argument: push_to or resync_destructive").
var ShclusterReplicatedConfigResyncCmd = newShclusterReplicatedConfigResyncCmd()
