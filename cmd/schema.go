package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandInfo describes one item in the schema output. Most items describe a
// runnable command; the top-level item describes flags shared by all commands.
type CommandInfo struct {
	Command  string     `json:"command"`
	Group    string     `json:"group"`
	Short    string     `json:"short"`
	Long     string     `json:"long,omitempty"`
	Examples string     `json:"examples,omitempty"`
	Flags    []FlagInfo `json:"flags,omitempty"`
}

// FlagInfo describes a single flag.
type FlagInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
	Usage   string `json:"usage,omitempty"`
	Env     string `json:"env,omitempty"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print full command tree as JSON (useful for LLM agents)",
	Long: `Emit a JSON array describing every available command, its flags, and examples.

An LLM agent can run this once to bootstrap its knowledge of splunkctl.

Use --compact to get a smaller payload (names and types only, no help text).
Use --groups to list available group names.
Use --group <name> to filter to a specific group.
The splunkctl entry lists global flags inherited by every command.`,
	Annotations: map[string]string{
		annotationNoClient: "true",
	},
	Example: `  splunkctl schema
  splunkctl schema --compact
  splunkctl schema --groups
  splunkctl schema --group cluster
  splunkctl schema --group search --compact
  splunkctl schema | jq '.[] | select(.command | startswith("index"))'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showGroups, _ := cmd.Flags().GetBool("groups")
		groupFilter, _ := cmd.Flags().GetString("group")
		compact, _ := cmd.Flags().GetBool("compact")

		var entries []CommandInfo
		collectCommands(cmd.Root(), "", &entries)

		if showGroups {
			return printGroups(entries)
		}

		if groupFilter != "" {
			filtered := make([]CommandInfo, 0, len(entries))
			for _, e := range entries {
				if e.Group == groupFilter {
					filtered = append(filtered, e)
				}
			}
			if len(filtered) == 0 {
				return fmt.Errorf("no commands found for group %q — run 'splunkctl schema --groups' to see available groups", groupFilter)
			}
			entries = filtered
		}

		if compact {
			entries = compactEntries(entries)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	},
}

func printGroups(entries []CommandInfo) error {
	seen := map[string]bool{}
	var groups []string
	for _, e := range entries {
		if !seen[e.Group] {
			seen[e.Group] = true
			groups = append(groups, e.Group)
		}
	}
	sort.Strings(groups)
	for _, g := range groups {
		fmt.Println(g)
	}
	return nil
}

func compactEntries(entries []CommandInfo) []CommandInfo {
	out := make([]CommandInfo, len(entries))
	for i, e := range entries {
		compacted := CommandInfo{
			Command: e.Command,
			Group:   e.Group,
			Short:   e.Short,
			// Long and Examples omitted
		}
		for _, f := range e.Flags {
			compacted.Flags = append(compacted.Flags, FlagInfo{
				Name: f.Name,
				Type: f.Type,
				Env:  f.Env,
				// Default and Usage omitted
			})
		}
		out[i] = compacted
	}
	return out
}

// groupFor returns the logical group name for a given object name (second word of the command).
func groupFor(objectName string) string {
	overrides := map[string]string{
		"shc-status":   "shcluster",
		"shc-upgrade":  "shcluster",
		"shc-recovery": "shcluster",

		"jobs":                     "search",
		"saved-search":             "search",
		"scheduler":                "search",
		"scheduler-status":         "search",
		"search-admission-control": "search",

		"role":                "auth",
		"user":                "auth",
		"licenser-groups":     "auth",
		"licenser-localpeer":  "auth",
		"licenser-messages":   "auth",
		"licenser-peers":      "auth",
		"licenser-pools":      "auth",
		"licenser-stacks":     "auth",
		"licenser-localslave": "auth",
		"licenses":            "auth",

		"monitor":         "inputs",
		"tcp":             "inputs",
		"udp":             "inputs",
		"exec":            "inputs",
		"scripted":        "inputs",
		"tail":            "inputs",
		"watch":           "inputs",
		"fifo":            "inputs",
		"wmi":             "inputs",
		"perfmon":         "inputs",
		"eventlog":        "inputs",
		"regmon":          "inputs",
		"winprintmon":     "inputs",
		"winhostmon":      "inputs",
		"winnetmon":       "inputs",
		"ad":              "inputs",
		"oneshot":         "inputs",
		"spool":           "inputs",
		"monitornohandle": "inputs",
		"inputstatus":     "inputs",

		"forward-server": "deployment",
		"deploy-client":  "deployment",
		"deploy-server":  "deployment",
		"deploy-poll":    "deployment",
		"serverclass":    "deployment",

		"app": "apps",

		"bundle-replication-config": "cluster",
		"bundle-replication-status": "cluster",
		"excess-buckets":            "cluster",
		"peer-buckets":              "cluster",
		"peer-info":                 "cluster",
		"maintenance-mode":          "cluster",

		"indexer-discovery": "index",
		"indexing-ready":    "index",
		"index":             "index",

		"workload-category":          "workload",
		"workload-config":            "workload",
		"workload-management":        "workload",
		"workload-management-status": "workload",
		"workload-policy":            "workload",
		"workload-pool":              "workload",
		"workload-rule":              "workload",

		"kvstore":                     "system",
		"kvstore-status":              "system",
		"kvstore-port":                "system",
		"kvstore-maintenance-mode":    "system",
		"health":                      "system",
		"log-level":                   "system",
		"splunkd":                     "system",
		"splunkd-port":                "system",
		"splunkweb":                   "system",
		"web-port":                    "system",
		"web-ssl":                     "system",
		"webserver":                   "system",
		"splunk-secret":               "system",
		"fips-mode":                   "system",
		"guid":                        "system",
		"servername":                  "system",
		"datastore-dir":               "system",
		"default-hostname":            "system",
		"minfreemb":                   "system",
		"boot-start":                  "system",
		"appserver-ports":             "system",
		"dfsmaster-port":              "system",
		"settings":                    "system",
		"server-status":               "system",
		"standalone":                  "system",
		"standalone-kvupgrade-status": "system",
		"postgres":                    "system",
		"postgres-health":             "system",
		"postgres-recovery-status":    "system",
		"postgres-status":             "system",
		"crl":                         "system",
		"listen":                      "system",
		"on":                          "system",
		"registry":                    "system",
		"manager-info":                "system",
		"remote-input-queue-config":   "system",
		"remote-input-queue-status":   "system",
		"remote-output-queue-config":  "system",
		"remote-output-queue-status":  "system",
		"cascade-plans":               "system",
		"cascade-plan-status":         "system",
		"version":                     "system",
		"schema":                      "system",
		"config":                      "system",
		"completion":                  "system",
		"help":                        "system",
	}
	if g, ok := overrides[objectName]; ok {
		return g
	}
	// default: first segment before first hyphen
	if idx := strings.Index(objectName, "-"); idx > 0 {
		return objectName[:idx]
	}
	return objectName
}

func collectCommands(cmd *cobra.Command, prefix string, out *[]CommandInfo) {
	name := cmd.Name()
	if prefix != "" {
		name = prefix + " " + cmd.Name()
	}

	// Flags such as --host, --token, --yes, and --read-only are defined once on
	// the top-level "splunkctl" command and are available to every subcommand.
	// Repeating them under every command would make the schema much larger, so
	// record them once while visiting the top-level command.
	if cmd == cmd.Root() {
		info := CommandInfo{
			Command: name,
			Group:   "system",
			Short:   "Global flags inherited by every command",
		}

		// Run this function once for each flag defined on the top-level command.
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			envName := ""
			// Host and token may also be supplied through SPLUNKCTL_HOST and
			// SPLUNKCTL_TOKEN. The other global flags do not have environment
			// variable alternatives, so their Env value remains empty.
			if f.Name == "host" || f.Name == "token" {
				envName = "SPLUNKCTL_" + strings.ToUpper(f.Name)
			}
			info.Flags = append(info.Flags, FlagInfo{
				Name:    f.Name,
				Type:    f.Value.Type(),
				Default: f.DefValue,
				Usage:   f.Usage,
				Env:     envName,
			})
		})

		// Add the completed top-level entry to the schema result. The function
		// continues below and visits the subcommands afterward.
		*out = append(*out, info)
	}

	if cmd.Runnable() {
		// extract object name: "splunkctl index list" → "index"
		parts := strings.Fields(name)
		objectName := ""
		if len(parts) >= 2 {
			objectName = parts[1] // parts[0] is "splunkctl"
		}

		info := CommandInfo{
			Command:  name,
			Group:    groupFor(objectName),
			Short:    cmd.Short,
			Long:     cmd.Long,
			Examples: cmd.Example,
		}
		// The global flags were already added above. For this command, include only
		// flags defined directly on it. For example, "index add" includes --name
		// here but does not repeat the global --yes flag.
		cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
			envName := "SPLUNKCTL_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			info.Flags = append(info.Flags, FlagInfo{
				Name:    f.Name,
				Type:    f.Value.Type(),
				Default: f.DefValue,
				Usage:   f.Usage,
				Env:     envName,
			})
		})
		*out = append(*out, info)
	}

	for _, sub := range cmd.Commands() {
		if !sub.Hidden {
			collectCommands(sub, name, out)
		}
	}
}

func init() {
	schemaCmd.Flags().Bool("compact", false, "Compact output: names and flag types only, no help text (~4K tokens vs ~26K)")
	schemaCmd.Flags().Bool("groups", false, "List available group names and exit")
	schemaCmd.Flags().String("group", "", "Filter to commands in this group (see --groups for names)")
	rootCmd.AddCommand(schemaCmd)
}
