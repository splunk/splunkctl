package override

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// objectPaths maps --type values to Splunk REST path suffixes (under /services).
var objectPaths = map[string]string{
	"dashboard":        "/data/ui/views/",
	"saved-search":     "/saved/searches/",
	"macro":            "/admin/macros/",
	"eventtype":        "/saved/eventtypes/",
	"field-extraction": "/data/props/extractions/",
	"lookup":           "/data/transforms/lookups/",
	"data-model":       "/data/models/",
	"scripted-input":   "/data/inputs/script/",
}

// AclCmd is the top-level acl command. Registered in cmd/register_overrides.go.
var AclCmd = &cobra.Command{
	Use:   "acl",
	Short: "Show or set permissions (ACL) on any knowledge object",
}

var aclShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show ACL for a knowledge object",
	Example: `  splunkctl acl show ops-monitor --type dashboard
  splunkctl acl show "Error Spike Alert" --type saved-search
  splunkctl acl show my_macro --type macro`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		objType, _ := cmd.Flags().GetString("type")
		if objType == "" {
			return fmt.Errorf("--type is required")
		}
		app, _ := cmd.Flags().GetString("app")
		aclP, err := resolveAclPath(cmd.Context(), c, objType, app, name)
		if err != nil {
			return err
		}
		body, status, err := c.RawDoAbs(cmd.Context(), "GET", aclP+"?output_mode=json", nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("acl show failed (%d): %s", status, string(body))
		}
		result, err := parseACLResponse(body, name, objType)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), result, outFmt)
	},
}

var aclSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set ACL for a knowledge object",
	Example: `  splunkctl acl set ops-monitor --type dashboard --sharing app --read "*" --write admin
  splunkctl acl set "Error Spike Alert" --type saved-search --sharing app --read admin --read power
  splunkctl acl set my_macro --type macro --owner alice`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		objType, _ := cmd.Flags().GetString("type")
		if objType == "" {
			return fmt.Errorf("--type is required")
		}
		app, _ := cmd.Flags().GetString("app")

		form := url.Values{}
		if sharing, _ := cmd.Flags().GetString("sharing"); sharing != "" {
			form.Set("sharing", sharing)
		}
		if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
			form.Set("owner", owner)
		}
		if readers, _ := cmd.Flags().GetStringArray("read"); len(readers) > 0 {
			form.Set("perms.read", strings.Join(readers, ", "))
		}
		if writers, _ := cmd.Flags().GetStringArray("write"); len(writers) > 0 {
			form.Set("perms.write", strings.Join(writers, ", "))
		}
		if len(form) == 0 {
			return fmt.Errorf("specify at least one of --sharing, --owner, --read, --write")
		}

		postPath, err := resolveAclPath(cmd.Context(), c, objType, app, name)
		if err != nil {
			return err
		}

		body, status, err := c.RawDoAbs(cmd.Context(), "POST", postPath+"?output_mode=json", form)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("acl set failed (%d): %s", status, string(body))
		}
		result, err := parseACLResponse(body, name, objType)
		if err != nil {
			return output.Print(cmd.Context(), []map[string]any{{"name": name, "type": objType, "updated": true}}, "")
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), result, outFmt)
	},
}

// resolveAclPath fetches the object via /services (supports wildcard lookup),
// extracts the real owner+app from the ACL metadata, and returns the concrete
// /servicesNS/<owner>/<app>/... ACL path. Splunk rejects "-" wildcard on both
// GET and POST to the /acl sub-resource.
func resolveAclPath(ctx context.Context, c interface {
	RawDo(context.Context, string, string, url.Values) ([]byte, int, error)
}, objType, app, name string) (string, error) {
	suffix, ok := objectPaths[objType]
	if !ok {
		return "", fmt.Errorf("unknown --type %q", objType)
	}
	if app == "" {
		app = "search"
	}
	// /services/ supports wildcard resolution across namespaces
	fetchPath := suffix + url.PathEscape(name) + "?output_mode=json"
	body, status, err := c.RawDo(ctx, "GET", fetchPath, nil)
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", fmt.Errorf("object lookup failed (%d): %s", status, string(body))
	}
	var feed struct {
		Entry []struct {
			ACL struct {
				Owner string `json:"owner"`
				App   string `json:"app"`
			} `json:"acl"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil || len(feed.Entry) == 0 {
		return "", fmt.Errorf("cannot resolve namespace for %s %q", objType, name)
	}
	owner := feed.Entry[0].ACL.Owner
	resolvedApp := feed.Entry[0].ACL.App
	if owner == "" {
		owner = "nobody"
	}
	if resolvedApp == "" {
		resolvedApp = app
	}
	return fmt.Sprintf("/servicesNS/%s/%s%s%s/acl",
		url.PathEscape(owner), url.PathEscape(resolvedApp), suffix, url.PathEscape(name)), nil
}

func parseACLResponse(body []byte, name, objType string) ([]map[string]any, error) {
	var feed struct {
		Entry []struct {
			ACL map[string]any `json:"acl"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil || len(feed.Entry) == 0 {
		return nil, fmt.Errorf("parsing acl response: %w", err)
	}
	acl := feed.Entry[0].ACL
	acl["name"] = name
	acl["type"] = objType
	return []map[string]any{acl}, nil
}

func init() {
	types := "dashboard, saved-search, macro, eventtype, field-extraction, lookup, data-model, scripted-input"

	for _, cmd := range []*cobra.Command{aclShowCmd, aclSetCmd} {
		cmd.Flags().String("name", "", "object name (positional or --name)")
		cmd.Flags().String("type", "", "object type: "+types)
		cmd.Flags().String("app", "search", `app context`)
	}
	aclSetCmd.Flags().String("sharing", "", "sharing level: user, app, or global")
	aclSetCmd.Flags().String("owner", "", "new owner username")
	aclSetCmd.Flags().StringArray("read", nil, "role with read access (repeatable: --read admin --read power)")
	aclSetCmd.Flags().StringArray("write", nil, "role with write access (repeatable: --write admin)")

	AclCmd.AddCommand(aclShowCmd, aclSetCmd)
}
