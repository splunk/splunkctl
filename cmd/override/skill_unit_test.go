package override

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
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

func newSkillMutationTestCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("agent", "codex", "")
	cmd.Flags().Bool("global", true, "")
	cmd.Flags().Bool("local", false, "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.SetContext(ctx)
	return cmd
}

func TestSkillMutationsCheckBeforeFilesystemChange(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	denied := errors.New("denied")
	ctx := cmdctx.WithWriteCheck(context.Background(), func(context.Context) error { return denied })
	target := filepath.Join(home, ".agents", "skills", "splunkctl", "SKILL.md")

	t.Run("install", func(t *testing.T) {
		err := runSkillInstall(newSkillMutationTestCommand(ctx), nil)
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want denied", err)
		}
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("skill file was created: %v", err)
		}
	})

	t.Run("remove", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
		err := runSkillRemove(newSkillMutationTestCommand(ctx), nil)
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want denied", err)
		}
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("skill file was removed: %v", err)
		}
	})
}
