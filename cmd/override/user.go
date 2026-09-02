package override

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// UserAddCmd overrides the generated add to accept --password and --role via --param too.
var UserAddCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Add a new user",
	Example: `  splunkctl user add jdoe --password "ch@ng3m#" --role power --full-name "John Doe"
  splunkctl user add jdoe --param password=TempPass123! --param roles=power --param email=jdoe@example.com`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		if username == "" && len(args) > 0 {
			username = args[0]
		}
		params := cmdctx.ExtraParams(cmd, map[string]string{"name": username})
		// --role / --password flags take precedence over --param equivalents
		if v, _ := cmd.Flags().GetString("role"); v != "" {
			params["roles"] = v
		}
		if v, _ := cmd.Flags().GetString("password"); v != "" {
			params["password"] = v
		}
		if v, _ := cmd.Flags().GetString("full-name"); v != "" {
			params["realname"] = v
		}
		if v, _ := cmd.Flags().GetString("tz"); v != "" {
			params["tz"] = v
		}
		if params["password"] == "" {
			return fmt.Errorf("password is required: use --password or --param password=<value>")
		}
		if params["roles"] == "" {
			return fmt.Errorf("role is required: use --role or --param roles=<value>")
		}
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}
		data, _, err := c.RawDo(cmd.Context(), "POST", "/authentication/users/", form)
		if err != nil {
			return err
		}
		results, err := parseUserFeed(data)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

// UserEditCmd overrides the generated edit to support multiple --roles values and --password.
var UserEditCmd = &cobra.Command{
	Use:   "edit <username>",
	Short: "Edit a user's roles, password, full name, or timezone",
	Example: `  splunkctl user edit alice --roles admin --roles can_delete
  splunkctl user edit alice --password 'NewSecure456!'
  splunkctl user edit alice --roles power --full-name "Alice Smith"
  splunkctl user edit alice --tz Europe/London`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		if username == "" && len(args) > 0 {
			username = args[0]
		}
		form := url.Values{}
		if roles, _ := cmd.Flags().GetStringArray("roles"); len(roles) > 0 {
			for _, r := range roles {
				form.Add("roles", r)
			}
		}
		if v, _ := cmd.Flags().GetString("password"); v != "" {
			form.Set("password", v)
		}
		if v, _ := cmd.Flags().GetString("full-name"); v != "" {
			form.Set("realname", v)
		}
		if v, _ := cmd.Flags().GetString("tz"); v != "" {
			form.Set("tz", v)
		}
		// Also pick up any --param values (e.g. --param password=X)
		extra := cmdctx.ExtraParams(cmd, map[string]string{})
		for k, v := range extra {
			form.Set(k, v)
		}
		if len(form) == 0 {
			return fmt.Errorf("specify at least one of --roles, --password, --full-name, --tz")
		}
		data, _, err := c.RawDo(cmd.Context(), "POST", "/authentication/users/"+username, form)
		if err != nil {
			return err
		}
		results, err := parseUserFeed(data)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

// UserRemoveCmd overrides the generated remove to return a clean confirmation.
var UserRemoveCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove a user",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		if username == "" && len(args) > 0 {
			username = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   "/authentication/users/",
			ID:     username,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		result := []map[string]any{{"username": username, "deleted": true}}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), result, outFmt)
	},
}

// UserPasswordCmd is injected into the generated UserCmd in cmd/register_overrides.go.
var UserPasswordCmd = &cobra.Command{
	Use:   "password <username>",
	Short: "Reset a user's password",
	Example: `  splunkctl user password jdoe --password newSecret123
  splunkctl user password --username jdoe --password newSecret123`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		username, _ := cmd.Flags().GetString("username")
		if username == "" && len(args) > 0 {
			username = args[0]
		}
		password, _ := cmd.Flags().GetString("password")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   "/authentication/users/",
			ID:     username,
			Params: cmdctx.ExtraParams(cmd, map[string]string{"password": password}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func parseUserFeed(body []byte) ([]map[string]any, error) {
	var feed struct {
		Entry []struct {
			Name    string         `json:"name"`
			Content map[string]any `json:"content"`
		} `json:"entry"`
		Messages []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parsing user response: %w", err)
	}
	if len(feed.Entry) == 0 && len(feed.Messages) > 0 {
		return nil, fmt.Errorf("splunk error: %s", feed.Messages[0].Text)
	}
	results := make([]map[string]any, 0, len(feed.Entry))
	for _, e := range feed.Entry {
		row := map[string]any{"name": e.Name}
		for k, v := range e.Content {
			row[k] = v
		}
		results = append(results, row)
	}
	return results, nil
}

func init() {
	UserAddCmd.Flags().String("username", "", "username (positional or --username)")
	UserAddCmd.Flags().String("role", "", `Role: Admin, Power, or User (or use --param roles=<value>)`)
	UserAddCmd.Flags().String("password", "", `Password (or use --param password=<value>)`)
	UserAddCmd.Flags().String("full-name", "", `Real name in quotes (e.g. "John Doe")`)
	UserAddCmd.Flags().String("tz", "", `Timezone (e.g. "Europe/London")`)

	UserEditCmd.Flags().StringArray("roles", nil, `Role(s) to assign (repeatable: --roles admin --roles can_delete)`)
	UserEditCmd.Flags().String("password", "", "new password")
	UserEditCmd.Flags().String("full-name", "", `Real name in quotes (e.g. "John Doe")`)
	UserEditCmd.Flags().String("tz", "", `Timezone (e.g. "Europe/London")`)
	UserEditCmd.Flags().String("username", "", "username identifier")

	UserRemoveCmd.Flags().String("username", "", "username identifier")

	UserPasswordCmd.Flags().String("username", "", "username (positional or --username)")
	UserPasswordCmd.Flags().String("password", "", "new password (required)")
	_ = UserPasswordCmd.MarkFlagRequired("password")
}
