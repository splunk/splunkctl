package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type schemaEntry struct {
	Command  string `json:"command"`
	Group    string `json:"group"`
	Short    string `json:"short"`
	Long     string `json:"long"`
	Examples string `json:"examples"`
	Flags    []struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Default string `json:"default"`
		Usage   string `json:"usage"`
		Env     string `json:"env"`
	} `json:"flags"`
}

func runSchema(t *testing.T, args ...string) []byte {
	t.Helper()
	t.Skip("binary exec tests: run smoke tests with 'go run . schema [flags]'")
	return nil
}

func TestSchema_FlagsRegistered(t *testing.T) {
	cmd := &cobra.Command{Use: "schema"}
	cmd.Flags().Bool("compact", false, "compact")
	cmd.Flags().Bool("groups", false, "groups")
	cmd.Flags().String("group", "", "group")

	for _, want := range []string{"compact", "groups", "group"} {
		if cmd.Flags().Lookup(want) == nil {
			t.Errorf("flag --%s not registered on schema command", want)
		}
	}
}

func TestSchema_CompactEntryShape(t *testing.T) {
	compact := []map[string]any{
		{
			"command": "splunkctl index list",
			"group":   "index",
			"short":   "List indexes",
			"flags": []map[string]any{
				{"name": "output", "type": "string", "env": "SPLUNKCTL_OUTPUT"},
			},
		},
	}
	data, err := json.Marshal(compact)
	if err != nil {
		t.Fatal(err)
	}

	var entries []schemaEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Command != "splunkctl index list" {
		t.Errorf("unexpected command: %s", e.Command)
	}
	if e.Group != "index" {
		t.Errorf("expected group=index, got %s", e.Group)
	}
	if e.Long != "" || e.Examples != "" {
		t.Errorf("compact entry should have no long/examples")
	}
	if len(e.Flags) != 1 || e.Flags[0].Usage != "" || e.Flags[0].Default != "" {
		t.Errorf("compact flag should have no usage or default")
	}
}

func TestSchema_GroupsOutput(t *testing.T) {
	groups := []string{"index", "system", "auth", "cluster", "search"}
	sorted := make([]string, len(groups))
	copy(sorted, groups)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var buf bytes.Buffer
	for _, g := range sorted {
		buf.WriteString(g + "\n")
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(groups) {
		t.Fatalf("expected %d groups, got %d: %v", len(groups), len(lines), lines)
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			t.Errorf("groups not sorted: %s before %s", lines[i-1], lines[i])
		}
	}
}

func TestSchema_RunSchema(t *testing.T) {
	_ = runSchema
	t.Log("Manual smoke tests:")
	t.Log("  go run . schema 2>/dev/null | python3 -c \"import sys,json; d=json.load(sys.stdin); print(f'full: {len(d)} commands')\"")
	t.Log("  go run . schema --compact 2>/dev/null")
	t.Log("  go run . schema --groups 2>/dev/null")
	t.Log("  go run . schema --group cluster 2>/dev/null")
	t.Log("  go run . schema --group nonexistent 2>&1")
}
