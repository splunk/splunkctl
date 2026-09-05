package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
)

type terminalBuffer struct {
	bytes.Buffer
	fd uintptr
}

func (b *terminalBuffer) Fd() uintptr { return b.fd }

type blockingTerminal struct {
	started  chan struct{}
	release  chan struct{}
	finished chan struct{}
	start    sync.Once
	finish   sync.Once
}

func (r *blockingTerminal) Read(p []byte) (int, error) {
	r.start.Do(func() { close(r.started) })
	<-r.release
	r.finish.Do(func() { close(r.finished) })
	return copy(p, "yes\n"), nil
}

func (r *blockingTerminal) Fd() uintptr { return 1 }

func newWriteCheckTestCommand(t *testing.T, yes, readOnly bool, input string, terminal bool) (*cobra.Command, client.WriteCheck, *terminalBuffer) {
	t.Helper()

	root := &cobra.Command{Use: "splunkctl"}
	object := &cobra.Command{Use: "index"}
	command := &cobra.Command{Use: "add"}
	root.AddCommand(object)
	object.AddCommand(command)
	command.Flags().Bool(flagYesName, yes, "")
	command.Flags().Bool(flagReadOnlyName, readOnly, "")

	in := &terminalBuffer{fd: 1}
	in.WriteString(input)
	errOut := &terminalBuffer{fd: 2}
	command.SetIn(in)
	command.SetErr(errOut)
	command.SetContext(context.Background())

	check, err := newWriteCheckWithTerminal(command, func(int) bool { return terminal })
	if err != nil {
		t.Fatal(err)
	}
	return command, check, errOut
}

func TestWriteCheckPolicy(t *testing.T) {
	tests := []struct {
		name     string
		yes      bool
		readOnly bool
		input    string
		terminal bool
		wantCode string
	}{
		{name: "yes skips prompt", yes: true},
		{name: "read only blocks", readOnly: true, wantCode: "READ_ONLY"},
		{name: "read only wins over yes", yes: true, readOnly: true, wantCode: "READ_ONLY"},
		{name: "agent requires approval", wantCode: "CONFIRMATION_REQUIRED"},
		{name: "interactive y", input: "y\n", terminal: true},
		{name: "interactive yes", input: "YES\n", terminal: true},
		{name: "interactive no", input: "n\n", terminal: true, wantCode: "CONFIRMATION_DECLINED"},
		{name: "interactive invalid", input: "maybe\n", terminal: true, wantCode: "CONFIRMATION_DECLINED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, check, errOut := newWriteCheckTestCommand(t, tt.yes, tt.readOnly, tt.input, tt.terminal)
			err := check(cmd.Context())
			if tt.wantCode == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var safetyErr *writeSafetyError
				if !errors.As(err, &safetyErr) || safetyErr.code != tt.wantCode {
					t.Fatalf("error = %v, want safety code %s", err, tt.wantCode)
				}
			}
			if !tt.terminal && errOut.Len() != 0 {
				t.Fatalf("non-terminal prompt = %q, want none", errOut.String())
			}
		})
	}
}

func TestWriteCheckRunsOnce(t *testing.T) {
	cmd, check, errOut := newWriteCheckTestCommand(t, false, false, "yes\nno\n", true)
	if err := check(cmd.Context()); err != nil {
		t.Fatal(err)
	}
	if err := check(cmd.Context()); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(errOut.String(), "Continue?"); got != 1 {
		t.Fatalf("prompt count = %d, want 1", got)
	}
}

func TestWriteCheckHonorsCanceledContext(t *testing.T) {
	for _, yes := range []bool{false, true} {
		cmd, check, _ := newWriteCheckTestCommand(t, yes, false, "yes\n", true)
		ctx, cancel := context.WithCancel(cmd.Context())
		cancel()
		if err := check(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("yes=%t: error = %v, want context canceled", yes, err)
		}
	}
}

func TestWriteCheckHonorsCancellationDuringPrompt(t *testing.T) {
	cmd, _, _ := newWriteCheckTestCommand(t, false, false, "", true)
	in := &blockingTerminal{
		started:  make(chan struct{}),
		release:  make(chan struct{}),
		finished: make(chan struct{}),
	}
	release := sync.OnceFunc(func() { close(in.release) })
	t.Cleanup(release)
	cmd.SetIn(in)
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()
	check, err := newWriteCheckWithTerminal(cmd, func(int) bool { return true })
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- check(ctx) }()
	select {
	case <-in.started:
	case <-time.After(5 * time.Second):
		t.Fatal("confirmation read did not start")
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("write check did not return after cancellation")
	}

	release()
	select {
	case <-in.finished:
	case <-time.After(5 * time.Second):
		t.Fatal("confirmation reader did not finish after release")
	}
	if err := check(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("late input changed cached result to %v, want context canceled", err)
	}
}

func TestRootWriteSafetyWiring(t *testing.T) {
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	remoteLeaf := &cobra.Command{
		Use:    "write-safety-test",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _, err := cmdctx.ClientFrom(cmd).RawDo(cmd.Context(), http.MethodPost, "/write-safety-test", nil)
			return err
		},
	}
	remoteLeaf.SetIn(&bytes.Buffer{})
	remoteLeaf.SetOut(io.Discard)
	remoteLeaf.SetErr(io.Discard)

	localEffects := 0
	localLeaf := &cobra.Command{
		Use:         "local-write-safety-test",
		Hidden:      true,
		Annotations: map[string]string{annotationNoClient: "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cmdctx.CheckWrite(cmd.Context()); err != nil {
				return err
			}
			localEffects++
			return nil
		},
	}
	rootCmd.AddCommand(remoteLeaf, localLeaf)

	oldContext := rootCmd.Context()
	type savedFlag struct {
		name    string
		value   string
		changed bool
	}
	var saved []savedFlag
	for _, name := range []string{"host", "token", "insecure", "output", flagYesName, flagReadOnlyName} {
		flag := rootCmd.PersistentFlags().Lookup(name)
		saved = append(saved, savedFlag{name: name, value: flag.Value.String(), changed: flag.Changed})
	}
	t.Cleanup(func() {
		rootCmd.RemoveCommand(remoteLeaf, localLeaf)
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(oldContext)
		for _, state := range saved {
			flag := rootCmd.PersistentFlags().Lookup(state.name)
			if err := flag.Value.Set(state.value); err != nil {
				t.Errorf("restore --%s: %v", state.name, err)
			}
			flag.Changed = state.changed
		}
	})
	tests := []struct {
		name     string
		yes      bool
		readOnly bool
		wantCode string
		wantSent int
	}{
		{name: "confirmation required", wantCode: "CONFIRMATION_REQUIRED"},
		{name: "read only", readOnly: true, wantCode: "READ_ONLY"},
		{name: "approved", yes: true, wantSent: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests = 0
			rootCmd.SetContext(context.Background())
			remoteLeaf.SetContext(nil)
			rootCmd.SetArgs([]string{
				"--host", server.URL,
				"--token", "test-token",
				"--insecure=true",
				"--output=json",
				fmt.Sprintf("--%s=%t", flagYesName, tt.yes),
				fmt.Sprintf("--%s=%t", flagReadOnlyName, tt.readOnly),
				remoteLeaf.Name(),
			})
			_, err := rootCmd.ExecuteC()

			if tt.wantCode == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var safetyErr *writeSafetyError
				if !errors.As(err, &safetyErr) || safetyErr.code != tt.wantCode {
					t.Fatalf("error = %v, want safety code %s", err, tt.wantCode)
				}
			}
			if requests != tt.wantSent {
				t.Fatalf("requests = %d, want %d", requests, tt.wantSent)
			}
		})
	}

	t.Run("no-client local command", func(t *testing.T) {
		localEffects = 0
		rootCmd.SetContext(context.Background())
		localLeaf.SetContext(nil)
		rootCmd.SetArgs([]string{
			fmt.Sprintf("--%s=true", flagYesName),
			fmt.Sprintf("--%s=false", flagReadOnlyName),
			localLeaf.Name(),
		})
		if _, err := rootCmd.ExecuteC(); err != nil {
			t.Fatal(err)
		}
		if localEffects != 1 {
			t.Fatalf("local effects = %d, want 1", localEffects)
		}
	})
}

func TestConfigWriteChecksBeforeFilesystemChange(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	denied := errors.New("denied")
	ctx := cmdctx.WithWriteCheck(context.Background(), func(context.Context) error { return denied })

	if err := writeConfig(ctx, "host", "https://splunk.example.com:8089"); !errors.Is(err, denied) {
		t.Fatalf("error = %v, want denied", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".splunkctl")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config directory was created: %v", err)
	}
}
