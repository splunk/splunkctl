package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const credentialPath = "/storage/passwords/"

// CredentialCmd is the top-level credential command. Registered in cmd/register_overrides.go.
var CredentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "Manage Splunk credential store (storage/passwords)",
}

var credentialListCmd = &cobra.Command{
	Use:         "list",
	Short:       "List stored credentials (passwords masked)",
	Example:     `  splunkctl credential list`,
	Annotations: map[string]string{"describe_path": credentialPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, credentialPath); done {
			return err
		}
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   credentialPath,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		// mask clear_password field for safety
		for _, r := range results {
			if _, ok := r["clear_password"]; ok {
				r["clear_password"] = "********"
			}
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var credentialAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Store a credential (username + password + realm)",
	Example: `  splunkctl credential add --username myuser --password mysecret --realm myapp
  splunkctl credential add --username apikey --password token123 --realm my_integration`,
	Annotations: map[string]string{"describe_path": credentialPath},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		if done, err := cmdctx.RunDescribe(cmd, c, credentialPath); done {
			return err
		}
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		realm, _ := cmd.Flags().GetString("realm")
		params := cmdctx.ExtraParams(cmd, map[string]string{
			"name":     username,
			"password": password,
			"realm":    realm,
		})
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   credentialPath,
			Params: params,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var credentialEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Update a stored credential's password",
	Example: `  splunkctl credential edit --username myuser --realm myapp --password newpassword`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		realm, _ := cmd.Flags().GetString("realm")
		// credential ID is realm:username:
		id := realm + ":" + username + ":"
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   credentialPath,
			ID:     id,
			Params: cmdctx.ExtraParams(cmd, map[string]string{"password": password}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var credentialRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Delete a stored credential",
	Example: `  splunkctl credential remove --username myuser --realm myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		realm, _ := cmd.Flags().GetString("realm")
		id := realm + ":" + username + ":"
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   credentialPath,
			ID:     id,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"username": username, "realm": realm, "deleted": true}}, outFmt)
	},
}

func init() {
	for _, cmd := range []*cobra.Command{credentialAddCmd, credentialEditCmd, credentialRemoveCmd} {
		cmd.Flags().String("username", "", "credential username (required)")
		cmd.Flags().String("realm", "", "credential realm/app context (required)")
		_ = cmd.MarkFlagRequired("username")
		_ = cmd.MarkFlagRequired("realm")
	}
	credentialAddCmd.Flags().String("password", "", "credential password (required)")
	credentialEditCmd.Flags().String("password", "", "new password (required)")
	_ = credentialAddCmd.MarkFlagRequired("password")
	_ = credentialEditCmd.MarkFlagRequired("password")
	CredentialCmd.AddCommand(credentialListCmd, credentialAddCmd, credentialEditCmd, credentialRemoveCmd)
}
