package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/output"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage splunkctl configuration",
	Annotations: map[string]string{
		annotationNoClient: "true",
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value (host or token)",
	Args:  cobra.ExactArgs(2),
	Annotations: map[string]string{
		annotationNoClient: "true",
	},
	Example: `  splunkctl config set host https://splunk.example.com:8089
  splunkctl config set token myapitoken`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if key != "host" && key != "token" {
			return fmt.Errorf("unknown config key %q: must be 'host' or 'token'", key)
		}
		return writeConfig(key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show current configuration",
	Annotations: map[string]string{
		annotationNoClient: "true",
	},
	Example: `  splunkctl config get`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		row := map[string]any{}
		for k, v := range cfg {
			if k == "token" {
				row[k] = "[REDACTED]"
			} else {
				row[k] = v
			}
		}
		return output.Print(cmd.Context(), []map[string]any{row}, flagOutput)
	},
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".splunkctl", "config.yaml"), nil
}

func readConfig() (map[string]string, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg map[string]string
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg == nil {
		cfg = make(map[string]string)
	}
	return cfg, nil
}

func writeConfig(key, value string) error {
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	cfg[key] = value

	path, err := configFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
