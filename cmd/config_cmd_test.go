package cmd_test

import (
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/config"
)

func TestConfig_MissingHostReturnsError(t *testing.T) {
	config.ResetForTesting()
	t.Setenv("SPLUNKCTL_HOST", "")
	t.Setenv("SPLUNKCTL_TOKEN", "")

	_, err := config.Resolve("", "", false, "")
	if err == nil {
		t.Fatal("expected error when no host provided")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Errorf("expected 'host' in error message, got: %v", err)
	}
}

func TestConfig_FlagBeatsEnvVar(t *testing.T) {
	config.Init()
	t.Setenv("SPLUNKCTL_HOST", "https://env-host:8089")
	t.Setenv("SPLUNKCTL_TOKEN", "env-token")

	cfg, err := config.Resolve("https://flag-host:8089", "flag-token", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://flag-host:8089" {
		t.Errorf("flag should win over env var, got %s", cfg.Host)
	}
	if cfg.Token != "flag-token" {
		t.Errorf("flag token should win over env var")
	}
}

func TestConfig_EnvVarUsedWhenNoFlag(t *testing.T) {
	config.Init()
	t.Setenv("SPLUNKCTL_HOST", "https://env-host:8089")
	t.Setenv("SPLUNKCTL_TOKEN", "env-token")

	cfg, err := config.Resolve("", "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://env-host:8089" {
		t.Errorf("expected env var host, got %s", cfg.Host)
	}
}
