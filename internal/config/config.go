package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	Token    string
	Insecure bool
	Output   string
}

// String redacts the token so Config is safe to log or format with %v.
func (c Config) String() string {
	return fmt.Sprintf("Config{Host:%s, Insecure:%v, Output:%s, Token:[REDACTED]}", c.Host, c.Insecure, c.Output)
}

// Init sets up Viper's env prefix and auto-env. Call once from main or root command init.
func Init() {
	viper.SetEnvPrefix("SPLUNKCTL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// config file: ~/.splunkctl/config.yaml
	if home, err := os.UserHomeDir(); err == nil {
		viper.SetConfigFile(filepath.Join(home, ".splunkctl", "config.yaml"))
		viper.SetConfigType("yaml")
		if err := viper.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) {
				fmt.Fprintf(os.Stderr, "warning: could not read config file: %v\n", err)
			}
		}
	}
}

// ResetForTesting clears Viper's in-memory state without re-reading any config file.
// Call this in tests that need to verify missing-value error paths.
func ResetForTesting() {
	viper.Reset()
	viper.SetEnvPrefix("SPLUNKCTL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	// intentionally skips config file so tests are not influenced by ~/.splunkctl/config.yaml
}

// Resolve returns a validated Config. Empty string values fall through to env vars.
// Resolution order: flag → SPLUNKCTL_* env var (via Viper).
func Resolve(flagHost, flagToken string, insecure bool, output string) (*Config, error) {
	host := firstNonEmpty(flagHost, viper.GetString("HOST"))
	token := firstNonEmpty(flagToken, viper.GetString("TOKEN"))

	if host == "" {
		return nil, errors.New("splunk host is required: set --host or SPLUNKCTL_HOST")
	}
	if token == "" {
		return nil, errors.New("splunk token is required: set --token or SPLUNKCTL_TOKEN")
	}

	host = normalizeHost(host)
	return &Config{Host: host, Token: token, Insecure: insecure, Output: output}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func normalizeHost(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	if strings.Contains(host, ":") {
		return fmt.Sprintf("https://%s", host)
	}
	return fmt.Sprintf("https://%s:8089", host)
}
