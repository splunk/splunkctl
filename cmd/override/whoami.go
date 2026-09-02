package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// WhoamiCmd is registered in cmd/register_overrides.go.
var WhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated identity, roles, and capabilities",
	Example: `  splunkctl whoami
  splunkctl whoami --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)

		// current-context gives us username + roles
		ctx, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/authentication/current-context/",
		})
		if err != nil {
			return err
		}

		// enrich with full user record (capabilities, email, etc.)
		if len(ctx) > 0 {
			username, _ := ctx[0]["username"].(string)
			if username != "" && username != "splunk-system-user" {
				users, err := c.Do(cmd.Context(), client.Request{
					Method: "GET",
					Path:   "/authentication/users/",
					ID:     username,
				})
				if err == nil && len(users) > 0 {
					// merge user record into context row
					for k, v := range users[0] {
						if _, exists := ctx[0][k]; !exists {
							ctx[0][k] = v
						}
					}
				}
			}
		}

		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), ctx, outFmt)
	},
}
