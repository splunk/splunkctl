package cmd_test

import (
	"regexp"
	"testing"

	"github.com/splunk/splunkctl/cmd"
)

func TestVersion_Defaults(t *testing.T) {
	// Default values set at build time when no ldflags are provided.
	if cmd.Version == "" {
		t.Error("Version should not be empty")
	}
	if cmd.Commit == "" {
		t.Error("Commit should not be empty")
	}
	if cmd.Built == "" {
		t.Error("Built should not be empty")
	}
}

func TestVersion_Format(t *testing.T) {
	semver := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	tests := []struct {
		input string
		valid bool
	}{
		{"0.1.0", true},
		{"1.0.0", true},
		{"10.20.30", true},
		{"0..a.1", false},
		{"v1.0.0", false},
		{"1.0", false},
		{"", false},
		{"dev", false},
	}
	for _, tc := range tests {
		if got := semver.MatchString(tc.input); got != tc.valid {
			t.Errorf("version %q: expected valid=%v, got %v", tc.input, tc.valid, got)
		}
	}
}

func TestVersion_DefaultOrSemver(t *testing.T) {
	semver := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if cmd.Version != "dev" && !semver.MatchString(cmd.Version) {
		t.Errorf("Version %q is not the dev placeholder or a valid semver (e.g. 0.1.0)", cmd.Version)
	}
}
