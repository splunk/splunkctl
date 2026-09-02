package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/config"
)

func TestMain(m *testing.M) {
	config.Init()
	os.Exit(m.Run())
}

func TestResolve_FlagsWin(t *testing.T) {
	os.Setenv("SPLUNKCTL_HOST", "https://env-host:8089")
	os.Setenv("SPLUNKCTL_TOKEN", "env-token")
	defer os.Unsetenv("SPLUNKCTL_HOST")
	defer os.Unsetenv("SPLUNKCTL_TOKEN")

	cfg, err := config.Resolve("https://flag-host:8089", "flag-token", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://flag-host:8089" {
		t.Errorf("expected flag host, got %s", cfg.Host)
	}
	if cfg.Token != "flag-token" {
		t.Errorf("expected flag-token, got <redacted>")
	}
}

func TestResolve_EnvFallback(t *testing.T) {
	os.Setenv("SPLUNKCTL_HOST", "https://env-host:8089")
	os.Setenv("SPLUNKCTL_TOKEN", "env-token")
	defer os.Unsetenv("SPLUNKCTL_HOST")
	defer os.Unsetenv("SPLUNKCTL_TOKEN")

	cfg, err := config.Resolve("", "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://env-host:8089" {
		t.Errorf("expected env host, got %s", cfg.Host)
	}
}

func TestResolve_MissingHost(t *testing.T) {
	os.Unsetenv("SPLUNKCTL_HOST")
	os.Unsetenv("SPLUNKCTL_TOKEN")
	config.ResetForTesting()

	_, err := config.Resolve("", "env-token", false, "")
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestResolve_MissingToken(t *testing.T) {
	os.Unsetenv("SPLUNKCTL_HOST")
	os.Unsetenv("SPLUNKCTL_TOKEN")

	_, err := config.Resolve("https://host:8089", "", false, "")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestResolve_BareHostGetsSchemeAndPort(t *testing.T) {
	cfg, err := config.Resolve("myhost", "tok", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://myhost:8089" {
		t.Errorf("expected normalized host, got %s", cfg.Host)
	}
}

func TestResolve_HostWithPortGetsScheme(t *testing.T) {
	cfg, err := config.Resolve("myhost:9999", "tok", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://myhost:9999" {
		t.Errorf("expected https://myhost:9999, got %s", cfg.Host)
	}
}

func TestResolve_AlreadyHTTPSUnchanged(t *testing.T) {
	cfg, err := config.Resolve("https://myhost:8089", "tok", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "https://myhost:8089" {
		t.Errorf("expected https://myhost:8089, got %s", cfg.Host)
	}
}

// String must not expose the token in logs or debug output.
func TestConfigString_RedactsToken(t *testing.T) {
	c := config.Config{Host: "https://myhost:8089", Token: "supersecrettoken"}
	got := c.String()
	if strings.Contains(got, "supersecrettoken") {
		t.Error("String() must not contain the real token")
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("String() must contain [REDACTED], got %q", got)
	}
}

// String must show the host so the config is still readable.
func TestConfigString_ShowsHost(t *testing.T) {
	c := config.Config{Host: "https://myhost:8089", Token: "supersecrettoken"}
	got := c.String()
	if !strings.Contains(got, "https://myhost:8089") {
		t.Errorf("String() must show the host, got %q", got)
	}
}
