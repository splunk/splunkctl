package override

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func parseDashboardFeed(body []byte) ([]map[string]any, error) {
	var feed struct {
		Entry []struct {
			Name    string         `json:"name"`
			Content map[string]any `json:"content"`
			ACL     map[string]any `json:"acl"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parsing dashboard response: %w", err)
	}
	results := make([]map[string]any, 0, len(feed.Entry))
	for _, e := range feed.Entry {
		row := map[string]any{"name": e.Name}
		for k, v := range e.Content {
			row[k] = v
		}
		if e.ACL != nil {
			row["sharing"] = e.ACL["sharing"]
			row["owner"] = e.ACL["owner"]
			row["app"] = e.ACL["app"]
		}
		results = append(results, row)
	}
	return results, nil
}

const dashboardSuffix = "/data/ui/views/"

// DashboardCmd is the top-level dashboard command. Registered in cmd/register_overrides.go.
var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Manage Splunk dashboards",
}

// dashboardPath returns the /services or /servicesNS path depending on whether app is set.
func dashboardPath(app string) string {
	if app != "" {
		return fmt.Sprintf("/servicesNS/nobody/%s%s", url.PathEscape(app), dashboardSuffix)
	}
	return "/services" + dashboardSuffix
}

var dashboardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dashboards",
	Example: `  splunkctl dashboard list
  splunkctl dashboard list --app search --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		path := dashboardPath(app)
		if name != "" {
			path += url.PathEscape(name)
		}
		body, status, err := c.RawDoAbs(cmd.Context(), "GET", path+"?output_mode=json", nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("dashboard list failed (%d): %s", status, string(body))
		}
		results, err := parseDashboardFeed(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var dashboardShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show dashboard definition (XML or JSON)",
	Example: `  splunkctl dashboard show soc-overview
  splunkctl dashboard show soc-overview --app search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		path := dashboardPath(app) + url.PathEscape(name) + "?output_mode=json"
		body, status, err := c.RawDoAbs(cmd.Context(), "GET", path, nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("dashboard show failed (%d): %s", status, string(body))
		}
		results, err := parseDashboardFeed(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var dashboardAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a dashboard from a file or inline content",
	Example: `  splunkctl dashboard add soc-overview --content '<dashboard><label>SOC Overview</label></dashboard>'
  splunkctl dashboard add soc-overview --file dashboard.xml --app search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		content, err := dashboardContent(cmd)
		if err != nil {
			return err
		}
		form := url.Values{}
		form.Set("name", name)
		form.Set("eai:data", content)
		for k, v := range cmdctx.ExtraParams(cmd, map[string]string{}) {
			form.Set(k, v)
		}
		path := dashboardPath(app) + "?output_mode=json"
		body, status, err := c.RawDoAbs(cmd.Context(), "POST", path, form)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("dashboard add failed (%d): %s", status, string(body))
		}
		results, err := parseDashboardFeed(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var dashboardEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Update a dashboard from a file or inline content",
	Example: `  splunkctl dashboard edit soc-overview --content '<dashboard><label>Updated</label></dashboard>'
  splunkctl dashboard edit soc-overview --file updated.xml --app search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		content, err := dashboardContent(cmd)
		if err != nil {
			return err
		}
		form := url.Values{}
		form.Set("eai:data", content)
		for k, v := range cmdctx.ExtraParams(cmd, map[string]string{}) {
			form.Set(k, v)
		}
		// Try the specified/default app path first; if 404, fall back to /services/ wildcard.
		path := dashboardPath(app) + url.PathEscape(name) + "?output_mode=json"
		body, status, err := c.RawDoAbs(cmd.Context(), "POST", path, form)
		if err != nil {
			return err
		}
		if status == 404 && app == "" {
			// Dashboard may have been created in a specific app — try resolving via /services/
			resolved, rerr := resolveDashboardPath(cmd.Context(), c, name)
			if rerr == nil {
				body, status, err = c.RawDoAbs(cmd.Context(), "POST", resolved+"?output_mode=json", form)
			}
		}
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("dashboard edit failed (%d): %s", status, string(body))
		}
		results, err := parseDashboardFeed(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var dashboardRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Delete a dashboard",
	Example: `  splunkctl dashboard remove soc-overview
  splunkctl dashboard remove soc-overview --app search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		path := dashboardPath(app) + url.PathEscape(name)
		_, status, err := c.RawDoAbs(cmd.Context(), "DELETE", path, nil)
		if err != nil {
			return err
		}
		if status == 404 && app == "" {
			resolved, rerr := resolveDashboardPath(cmd.Context(), c, name)
			if rerr == nil {
				_, status, err = c.RawDoAbs(cmd.Context(), "DELETE", resolved, nil)
			}
		}
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("dashboard remove failed (%d)", status)
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

// resolveDashboardPath fetches the dashboard via /services/ and returns the
// concrete /servicesNS/<owner>/<app>/ path for write operations.
func resolveDashboardPath(ctx context.Context, c *client.Client, name string) (string, error) {
	body, status, err := c.RawDo(ctx, "GET", dashboardSuffix+url.PathEscape(name)+"?output_mode=json", nil)
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", fmt.Errorf("dashboard lookup failed (%d)", status)
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
		return "", fmt.Errorf("cannot resolve namespace for dashboard %q", name)
	}
	owner := feed.Entry[0].ACL.Owner
	app := feed.Entry[0].ACL.App
	if owner == "" {
		owner = "nobody"
	}
	if app == "" {
		app = "search"
	}
	return fmt.Sprintf("/servicesNS/%s/%s%s%s",
		url.PathEscape(owner), url.PathEscape(app), dashboardSuffix, url.PathEscape(name)), nil
}

// dashboardContent returns eai:data from --content or --file, requiring exactly one.
func dashboardContent(cmd *cobra.Command) (string, error) {
	content, _ := cmd.Flags().GetString("content")
	if content != "" {
		return content, nil
	}
	filePath, _ := cmd.Flags().GetString("file")
	if filePath == "" {
		return "", fmt.Errorf("provide dashboard content via --content or --file")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func init() {
	for _, cmd := range []*cobra.Command{dashboardListCmd, dashboardShowCmd, dashboardAddCmd, dashboardEditCmd, dashboardRemoveCmd} {
		cmd.Flags().String("name", "", "dashboard name (positional or --name)")
		cmd.Flags().String("app", "", `app context (e.g. search)`)
	}
	for _, cmd := range []*cobra.Command{dashboardAddCmd, dashboardEditCmd} {
		cmd.Flags().String("file", "", "path to dashboard XML/JSON file")
		cmd.Flags().String("content", "", "dashboard XML/JSON content (inline alternative to --file)")
	}

	DashboardCmd.AddCommand(dashboardListCmd, dashboardShowCmd, dashboardAddCmd, dashboardEditCmd, dashboardRemoveCmd)
}
