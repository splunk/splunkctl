package override

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

//go:embed skill_claude.md skill_codex.md
var skillFS embed.FS

// SkillCmd is the top-level skill command.
var SkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the splunkctl agent skill for Claude Code and Codex",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the splunkctl agent skill into Claude Code or Codex",
	Long: `Install the splunkctl skill file so AI agents (Claude Code, Codex) know how to
use this CLI effectively.

Flags mode (non-interactive, suitable for scripts and agents):
  splunkctl skill install --agent claude --global
  splunkctl skill install --agent codex --local
  splunkctl skill install --agent both

Interactive mode (prompts for choices):
  splunkctl skill install --interactive`,
	Example: `  splunkctl skill install --interactive
  splunkctl skill install --agent claude --global
  splunkctl skill install --agent codex --local
  splunkctl skill install --agent both --global`,
	Annotations: map[string]string{"no_client": "true"},
	RunE:        runSkillInstall,
}

var skillListCmd = &cobra.Command{
	Use:         "list",
	Short:       "Show where the splunkctl skill is installed",
	Example:     `  splunkctl skill list`,
	Annotations: map[string]string{"no_client": "true"},
	RunE:        runSkillList,
}

var skillRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the splunkctl skill from Claude Code or Codex",
	Example: `  splunkctl skill remove --agent claude --global
  splunkctl skill remove --agent both`,
	Annotations: map[string]string{"no_client": "true"},
	RunE:        runSkillRemove,
}

func init() {
	for _, cmd := range []*cobra.Command{skillInstallCmd, skillRemoveCmd} {
		cmd.Flags().String("agent", "", "Agent to target: claude, codex, or both")
		cmd.Flags().Bool("global", false, "Install/remove at user-global scope (~/.claude or ~/.agents)")
		cmd.Flags().Bool("local", false, "Install/remove at project-local scope (.claude or .agents in CWD)")
		cmd.Flags().Bool("interactive", false, "Prompt for agent and scope interactively")
	}
	SkillCmd.AddCommand(skillInstallCmd, skillListCmd, skillRemoveCmd)
}

// skillTarget describes one installation target.
type skillTarget struct {
	Agent    string // "claude" or "codex"
	Scope    string // "global" or "local"
	Dir      string // resolved absolute directory path
	Filename string // SKILL.md
	Content  string // rendered skill content
}

func resolveTargets(agent, scope string) ([]skillTarget, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot determine working directory: %w", err)
	}

	claudeContent, err := skillFS.ReadFile("skill_claude.md")
	if err != nil {
		return nil, err
	}
	codexContent, err := skillFS.ReadFile("skill_codex.md")
	if err != nil {
		return nil, err
	}

	dirs := map[string]map[string]string{
		"claude": {
			"global": filepath.Join(home, ".claude", "skills", "splunkctl"),
			"local":  filepath.Join(cwd, ".claude", "skills", "splunkctl"),
		},
		"codex": {
			"global": filepath.Join(home, ".agents", "skills", "splunkctl"),
			"local":  filepath.Join(cwd, ".agents", "skills", "splunkctl"),
		},
	}
	contents := map[string]string{
		"claude": string(claudeContent),
		"codex":  string(codexContent),
	}

	agents := []string{}
	switch agent {
	case "claude", "codex":
		agents = []string{agent}
	case "both":
		agents = []string{"claude", "codex"}
	default:
		return nil, fmt.Errorf("unknown agent %q: must be claude, codex, or both", agent)
	}

	var targets []skillTarget
	for _, a := range agents {
		targets = append(targets, skillTarget{
			Agent:    a,
			Scope:    scope,
			Dir:      dirs[a][scope],
			Filename: "SKILL.md",
			Content:  contents[a],
		})
	}
	return targets, nil
}

func resolveScope(global, local bool) (string, error) {
	if global && local {
		return "", fmt.Errorf("--global and --local are mutually exclusive")
	}
	if global {
		return "global", nil
	}
	return "local", nil // default to local
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	agent, _ := cmd.Flags().GetString("agent")
	global, _ := cmd.Flags().GetBool("global")
	local, _ := cmd.Flags().GetBool("local")
	interactive, _ := cmd.Flags().GetBool("interactive")

	if interactive {
		var err error
		agent, global, err = promptInstallOptions()
		if err != nil {
			return err
		}
		local = !global
	} else {
		if agent == "" {
			return fmt.Errorf("--agent is required (claude, codex, or both) — or use --interactive")
		}
	}

	scope, err := resolveScope(global, local)
	if err != nil {
		return err
	}

	targets, err := resolveTargets(agent, scope)
	if err != nil {
		return err
	}
	// Installing a skill creates a directory and a SKILL.md file on this computer.
	// It does not send an HTTP request, so check permission before creating them.
	if err := cmdctx.CheckWrite(cmd.Context()); err != nil {
		return err
	}

	for _, t := range targets {
		if err := os.MkdirAll(t.Dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", t.Dir, err)
		}
		dest := filepath.Join(t.Dir, t.Filename)
		if err := os.WriteFile(dest, []byte(t.Content), 0644); err != nil {
			return fmt.Errorf("writing skill file %s: %w", dest, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Installed %s skill (%s) → %s\n", t.Agent, t.Scope, dest)
	}

	if agent == "claude" || agent == "both" {
		fmt.Fprintln(cmd.OutOrStdout(), "\nTo use in Claude Code: invoke the skill with /skill splunkctl or let Claude auto-detect it.")
	}
	if agent == "codex" || agent == "both" {
		fmt.Fprintln(cmd.OutOrStdout(), "\nTo use in Codex: type $splunkctl in your prompt to invoke the skill.")
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\nMake sure splunkctl is on your PATH so the agent can run it:")
	fmt.Fprintln(cmd.OutOrStdout(), "  cp splunkctl /usr/local/bin/          # install system-wide")
	fmt.Fprintln(cmd.OutOrStdout(), "  export PATH=\"$PATH:$(pwd)\"            # or add current dir to PATH")
	return nil
}

func runSkillList(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cwd, _ := os.Getwd()

	candidates := []map[string]any{
		{"agent": "claude", "scope": "global", "path": filepath.Join(home, ".claude", "skills", "splunkctl", "SKILL.md")},
		{"agent": "claude", "scope": "local", "path": filepath.Join(cwd, ".claude", "skills", "splunkctl", "SKILL.md")},
		{"agent": "codex", "scope": "global", "path": filepath.Join(home, ".agents", "skills", "splunkctl", "SKILL.md")},
		{"agent": "codex", "scope": "local", "path": filepath.Join(cwd, ".agents", "skills", "splunkctl", "SKILL.md")},
	}

	var found []map[string]any
	for _, c := range candidates {
		path := c["path"].(string)
		info, err := os.Stat(path)
		if err == nil {
			c["installed"] = true
			c["size"] = info.Size()
			c["modified"] = info.ModTime().Format("2006-01-02 15:04:05")
			found = append(found, c)
		}
	}

	if len(found) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No splunkctl skills installed. Run 'splunkctl skill install --interactive' to install.")
		return nil
	}

	outFmt, _ := cmd.Flags().GetString("output")
	return output.Print(cmd.Context(), found, outFmt)
}

func runSkillRemove(cmd *cobra.Command, args []string) error {
	agent, _ := cmd.Flags().GetString("agent")
	global, _ := cmd.Flags().GetBool("global")
	local, _ := cmd.Flags().GetBool("local")
	interactive, _ := cmd.Flags().GetBool("interactive")

	if interactive {
		var err error
		agent, global, err = promptInstallOptions()
		if err != nil {
			return err
		}
		local = !global
	} else {
		if agent == "" {
			return fmt.Errorf("--agent is required (claude, codex, or both) — or use --interactive")
		}
	}

	scope, err := resolveScope(global, local)
	if err != nil {
		return err
	}

	targets, err := resolveTargets(agent, scope)
	if err != nil {
		return err
	}
	// Removing a skill deletes a local SKILL.md file and possibly its empty directory.
	// It does not send an HTTP request, so check permission before deleting either.
	if err := cmdctx.CheckWrite(cmd.Context()); err != nil {
		return err
	}

	for _, t := range targets {
		dest := filepath.Join(t.Dir, t.Filename)
		err := os.Remove(dest)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s skill (%s) not found at %s\n", t.Agent, t.Scope, dest)
			continue
		}
		if err != nil {
			return fmt.Errorf("removing %s: %w", dest, err)
		}
		// remove parent dir if empty
		_ = os.Remove(t.Dir)
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed %s skill (%s) from %s\n", t.Agent, t.Scope, dest)
	}
	return nil
}

// promptInstallOptions asks the user for agent and scope interactively via stdin.
func promptInstallOptions() (agent string, global bool, err error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Which agent? [claude/codex/both] (default: both): ")
	scanner.Scan()
	agent = strings.TrimSpace(scanner.Text())
	if agent == "" {
		agent = "both"
	}
	if agent != "claude" && agent != "codex" && agent != "both" {
		return "", false, fmt.Errorf("invalid agent %q: must be claude, codex, or both", agent)
	}

	fmt.Print("Scope? [global/local] (global = ~/.claude or ~/.agents, local = ./.claude or ./.agents, default: global): ")
	scanner.Scan()
	scope := strings.TrimSpace(scanner.Text())
	if scope == "" {
		scope = "global"
	}
	if scope != "global" && scope != "local" {
		return "", false, fmt.Errorf("invalid scope %q: must be global or local", scope)
	}

	return agent, scope == "global", nil
}
