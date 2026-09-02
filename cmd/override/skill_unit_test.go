package override

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunSkillList_NoneInstalled(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.SetContext(context.Background())
	cmd.SetOut(&buf)

	err := runSkillList(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "No splunkctl skills installed") {
		t.Errorf("expected 'No splunkctl skills installed' in output, got %q", buf.String())
	}
}

func TestResolveTargets_UnknownAgent(t *testing.T) {
	_, err := resolveTargets("badagent", "global")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected 'unknown agent' in error, got %v", err)
	}
}

func TestResolveTargets_ClaudeGlobal(t *testing.T) {
	targets, err := resolveTargets("claude", "global")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	got := targets[0]
	if got.Agent != "claude" {
		t.Errorf("expected Agent=claude, got %q", got.Agent)
	}
	if got.Scope != "global" {
		t.Errorf("expected Scope=global, got %q", got.Scope)
	}
	if got.Filename != "SKILL.md" {
		t.Errorf("expected Filename=SKILL.md, got %q", got.Filename)
	}
	wantDirSuffix := filepath.Join(".claude", "skills", "splunkctl")
	if !strings.Contains(got.Dir, wantDirSuffix) {
		t.Errorf("expected Dir to contain %q, got %q", wantDirSuffix, got.Dir)
	}
}

func TestResolveTargets_CodexLocal(t *testing.T) {
	targets, err := resolveTargets("codex", "local")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	got := targets[0]
	if got.Agent != "codex" {
		t.Errorf("expected Agent=codex, got %q", got.Agent)
	}
	wantDirSuffix := filepath.Join(".agents", "skills", "splunkctl")
	if !strings.Contains(got.Dir, wantDirSuffix) {
		t.Errorf("expected Dir to contain %q, got %q", wantDirSuffix, got.Dir)
	}
	if got.Content == "" {
		t.Error("expected non-empty Content from embedded skill file")
	}
}

func TestResolveTargets_BothAgents(t *testing.T) {
	targets, err := resolveTargets("both", "global")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	agents := map[string]bool{}
	for _, tgt := range targets {
		agents[tgt.Agent] = true
	}
	if !agents["claude"] {
		t.Error("expected a target with Agent=claude")
	}
	if !agents["codex"] {
		t.Error("expected a target with Agent=codex")
	}
}

func TestResolveScope_BothFlagsSet(t *testing.T) {
	_, err := resolveScope(true, true)
	if err == nil {
		t.Fatal("expected error when both --global and --local are set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error, got %v", err)
	}
}

func TestResolveScope_GlobalOnly(t *testing.T) {
	got, err := resolveScope(true, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "global" {
		t.Errorf("got %q, want %q", got, "global")
	}
}

func TestResolveScope_DefaultToLocal(t *testing.T) {
	got, err := resolveScope(false, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "local" {
		t.Errorf("got %q, want %q", got, "local")
	}
}
