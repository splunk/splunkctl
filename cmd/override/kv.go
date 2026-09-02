package override

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// KvCmd is the top-level kv command. Registered in cmd/register_overrides.go.
var KvCmd = &cobra.Command{
	Use:   "kv",
	Short: "Manage KV Store collections and data",
}

// ── collection subcommand ────────────────────────────────────────────────────

var kvCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Manage KV Store collection definitions",
}

var kvCollectionListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List KV Store collections",
	Example: `  splunkctl kv collection list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		appFlag, _ := cmd.Flags().GetString("app")
		extra := cmdctx.ExtraParams(cmd, map[string]string{})
		app, _ := kvAppFromParams(appFlag, extra)
		path := kvCollectionPath(app) + "?output_mode=json"
		body, status, err := c.RawDoAbs(cmd.Context(), "GET", path, nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv collection list failed (%d): %s", status, string(body))
		}
		results, err := kvParseEntries(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var kvCollectionAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a KV Store collection",
	Example: `  splunkctl kv collection add my_collection --app search
  splunkctl kv collection add my_collection --app search --param field.id=number`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		extra := cmdctx.ExtraParams(cmd, map[string]string{})
		app, extra := kvAppFromParams(appFlag, extra)
		form := url.Values{"name": {name}}
		for k, v := range extra {
			form.Set(k, v)
		}
		body, status, err := c.RawDoAbs(cmd.Context(), "POST", kvCollectionPath(app)+"?output_mode=json", form)
		if err != nil {
			return err
		}
		if status == 409 {
			return fmt.Errorf("collection %q already exists (use 'kv collection list' to see existing collections)", name)
		}
		if status >= 400 {
			return fmt.Errorf("kv collection add failed (%d): %s", status, string(body))
		}
		results, err := kvParseEntries(body)
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var kvCollectionRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a KV Store collection",
	Example: `  splunkctl kv collection remove my_collection --app search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		extra := cmdctx.ExtraParams(cmd, map[string]string{})
		app, _ := kvAppFromParams(appFlag, extra)
		path := kvCollectionPath(app) + url.PathEscape(name) + "?output_mode=json"
		body, status, err := c.RawDoAbs(cmd.Context(), "DELETE", path, nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv collection remove failed (%d): %s", status, string(body))
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

// ── data subcommand ──────────────────────────────────────────────────────────

var kvDataCmd = &cobra.Command{
	Use:   "data",
	Short: "Read and write KV Store collection data",
}

var kvDataQueryCmd = &cobra.Command{
	Use:   "query <collection>",
	Short: "Query records from a KV Store collection",
	Example: `  splunkctl kv data query my_collection --app search
  splunkctl kv data query my_collection --app search --param query='{"status":"active"}'`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		collection, _ := cmd.Flags().GetString("collection")
		if collection == "" && len(args) > 0 {
			collection = args[0]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		params := cmdctx.ExtraParams(cmd, map[string]string{})
		app, params := kvAppFromParams(appFlag, params)

		body, status, err := c.RawDoAbs(cmd.Context(), "GET",
			kvDataPath(app, collection)+"?output_mode=json&"+encodeParams(params), nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv query failed (%d): %s", status, string(body))
		}

		var records []map[string]any
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parsing kv results: %w", err)
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), records, outFmt)
	},
}

var kvDataGetCmd = &cobra.Command{
	Use:     "get <collection> <key>",
	Short:   "Get a single record by _key from a KV Store collection",
	Example: `  splunkctl kv data get my_collection abc123 --app search`,
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		collection, _ := cmd.Flags().GetString("collection")
		key, _ := cmd.Flags().GetString("key")
		if collection == "" && len(args) > 0 {
			collection = args[0]
		}
		if key == "" && len(args) > 1 {
			key = args[1]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		app, _ := kvAppFromParams(appFlag, cmdctx.ExtraParams(cmd, map[string]string{}))

		body, status, err := c.RawDoAbs(cmd.Context(), "GET",
			kvDataPath(app, collection)+"/"+url.PathEscape(key)+"?output_mode=json", nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv get failed (%d): %s", status, string(body))
		}
		var record map[string]any
		if err := json.Unmarshal(body, &record); err != nil {
			return fmt.Errorf("parsing kv record: %w", err)
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{record}, outFmt)
	},
}

var kvDataPutCmd = &cobra.Command{
	Use:     "put <collection>",
	Short:   "Insert or update a JSON record in a KV Store collection",
	Example: `  splunkctl kv data put my_collection --app search --data '{"_key":"abc123","status":"active"}'`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		collection, _ := cmd.Flags().GetString("collection")
		if collection == "" && len(args) > 0 {
			collection = args[0]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		app, _ := kvAppFromParams(appFlag, cmdctx.ExtraParams(cmd, map[string]string{}))
		data, _ := cmd.Flags().GetString("data")

		// validate JSON
		var record map[string]any
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			return fmt.Errorf("--data must be valid JSON: %w", err)
		}

		// POST to base path — include _key in body for create-or-replace.
		// POSTing to /<key> would require the record to already exist.
		path := kvDataPath(app, collection)

		// KV Store data endpoint requires application/json body
		body, status, err := c.RawDoJSON(cmd.Context(), "POST", path, record)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv put failed (%d): %s", status, string(body))
		}
		var result map[string]any
		if err := json.Unmarshal(body, &result); err != nil {
			result = map[string]any{"ok": true}
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{result}, outFmt)
	},
}

var kvDataDeleteCmd = &cobra.Command{
	Use:     "delete <collection> <key>",
	Short:   "Delete a record by _key from a KV Store collection",
	Example: `  splunkctl kv data delete my_collection abc123 --app search`,
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		collection, _ := cmd.Flags().GetString("collection")
		key, _ := cmd.Flags().GetString("key")
		if collection == "" && len(args) > 0 {
			collection = args[0]
		}
		if key == "" && len(args) > 1 {
			key = args[1]
		}
		appFlag, _ := cmd.Flags().GetString("app")
		app, _ := kvAppFromParams(appFlag, cmdctx.ExtraParams(cmd, map[string]string{}))

		path := kvDataPath(app, collection) + "/" + url.PathEscape(key)
		body, status, err := c.RawDoAbs(cmd.Context(), "DELETE", path+"?output_mode=json", nil)
		if err != nil {
			return err
		}
		if status >= 400 {
			return fmt.Errorf("kv delete failed (%d): %s", status, string(body))
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"_key": key, "deleted": true}}, outFmt)
	},
}

// kvAppFromParams extracts "app" from extra params map (into the URL path, not POST body).
// Returns the resolved app name and the cleaned params map with "app" removed.
func kvAppFromParams(flagApp string, extra map[string]string) (string, map[string]string) {
	if v, ok := extra["app"]; ok {
		delete(extra, "app")
		if flagApp == "" {
			return v, extra
		}
	}
	return flagApp, extra
}

// kvParseEntries parses a Splunk Atom feed response into flat maps.
func kvParseEntries(body []byte) ([]map[string]any, error) {
	var feed struct {
		Entry []struct {
			Name    string         `json:"name"`
			Content map[string]any `json:"content"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parsing kv response: %w", err)
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

func kvCollectionPath(app string) string {
	if app != "" {
		return "/servicesNS/nobody/" + url.PathEscape(app) + "/storage/collections/config/"
	}
	return "/servicesNS/nobody/search/storage/collections/config/"
}

func kvDataPath(app, collection string) string {
	if app != "" {
		return "/servicesNS/nobody/" + url.PathEscape(app) + "/storage/collections/data/" + url.PathEscape(collection)
	}
	return "/servicesNS/nobody/search/storage/collections/data/" + url.PathEscape(collection)
}

func encodeParams(params map[string]string) string {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return v.Encode()
}

func init() {
	for _, cmd := range []*cobra.Command{kvCollectionListCmd, kvCollectionAddCmd, kvCollectionRemoveCmd} {
		cmd.Flags().String("app", "", `app context for the collection (default: "search")`)
	}
	kvCollectionAddCmd.Flags().String("name", "", "collection name (positional or --name)")
	kvCollectionRemoveCmd.Flags().String("name", "", "collection name (positional or --name)")

	for _, cmd := range []*cobra.Command{kvDataQueryCmd, kvDataGetCmd, kvDataPutCmd, kvDataDeleteCmd} {
		cmd.Flags().String("app", "", `app context for the collection (default: "search")`)
		cmd.Flags().String("collection", "", "collection name (positional or --collection)")
	}
	kvDataGetCmd.Flags().String("key", "", "_key of the record (positional or --key)")
	kvDataDeleteCmd.Flags().String("key", "", "_key of the record (positional or --key)")
	kvDataPutCmd.Flags().String("data", "", "JSON record to insert/update (required)")
	_ = kvDataPutCmd.MarkFlagRequired("data")

	kvCollectionCmd.AddCommand(kvCollectionListCmd, kvCollectionAddCmd, kvCollectionRemoveCmd)
	kvDataCmd.AddCommand(kvDataQueryCmd, kvDataGetCmd, kvDataPutCmd, kvDataDeleteCmd)
	KvCmd.AddCommand(kvCollectionCmd, kvDataCmd)
}
