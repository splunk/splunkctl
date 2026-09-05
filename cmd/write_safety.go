package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"golang.org/x/term"
)

const (
	flagYesName      = "yes"
	flagReadOnlyName = "read-only"
)

// writeSafetyError keeps both the message shown to a person and the short code
// returned in JSON, such as READ_ONLY or CONFIRMATION_REQUIRED. cause preserves
// an underlying input or output error when one exists.
type writeSafetyError struct {
	code    string
	message string
	cause   error
}

func (e *writeSafetyError) Error() string { return e.message }

func (e *writeSafetyError) Unwrap() error { return e.cause }

// newWriteCheck creates the function that decides whether a write may continue.
// term.IsTerminal tells it whether a person can answer a prompt in this terminal.
func newWriteCheck(cmd *cobra.Command) (client.WriteCheck, error) {
	return newWriteCheckWithTerminal(cmd, term.IsTerminal)
}

// Accepting the terminal check as an argument lets tests simulate both a person
// using a terminal and an agent running without an interactive terminal.
func newWriteCheckWithTerminal(cmd *cobra.Command, isTerminal func(int) bool) (client.WriteCheck, error) {
	// Copy the two flag values now. These values are used for every write made
	// by this command and do not change while the command is running.
	yes, err := cmd.Flags().GetBool(flagYesName)
	if err != nil {
		return nil, err
	}
	readOnly, err := cmd.Flags().GetBool(flagReadOnlyName)
	if err != nil {
		return nil, err
	}

	// One command can call the check more than once. For example, lookup upload
	// checks before creating a local file and again before sending its POST.
	// sync.Once runs authorizeWrite only the first time, and result remembers
	// whether that first check allowed or rejected the write.
	var once sync.Once
	var result error
	return func(ctx context.Context) error {
		// Check --read-only first. If both --read-only and --yes were supplied,
		// --read-only wins and the write is rejected.
		if readOnly {
			return &writeSafetyError{
				code:    "READ_ONLY",
				message: fmt.Sprintf("write operation for %q is blocked by --read-only", cmd.CommandPath()),
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		// The first call asks for confirmation or honors --yes. Later calls skip
		// this function body and reuse the answer stored in result.
		once.Do(func() {
			result = authorizeWrite(ctx, cmd, yes, isTerminal)
		})

		// The user may cancel the command while the prompt is waiting for input.
		// Check again so a cancellation always stops the request.
		if err := ctx.Err(); err != nil {
			return err
		}
		return result
	}, nil
}

func authorizeWrite(ctx context.Context, cmd *cobra.Command, yes bool, isTerminal func(int) bool) error {
	commandPath := cmd.CommandPath()
	// --yes means the user already approved writes when starting the command.
	if yes {
		return nil
	}

	in := cmd.InOrStdin()
	errOut := cmd.ErrOrStderr()
	// Ask a question only when both input and error output are attached to a real
	// terminal. An agent, script, or redirected command cannot answer safely, so
	// return an error that tells it to obtain approval and rerun with --yes.
	if !terminalStream(in, isTerminal) || !terminalStream(errOut, isTerminal) {
		return &writeSafetyError{
			code:    "CONFIRMATION_REQUIRED",
			message: fmt.Sprintf("write operation for %q requires confirmation; rerun with --yes after user approval", commandPath),
		}
	}

	if _, err := fmt.Fprintf(errOut, "WARNING: %q can modify state. Continue? [y/N]: ", commandPath); err != nil {
		return &writeSafetyError{code: "CONFIRMATION_ERROR", message: "write confirmation could not be displayed", cause: err}
	}
	answer, err := readLine(ctx, in)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return &writeSafetyError{code: "CONFIRMATION_ERROR", message: "write confirmation could not be read", cause: err}
	}
	// Remove spaces and the Enter key's newline, then ignore letter case. Only
	// "y" or "yes" allows the write; an empty or different answer rejects it.
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		return &writeSafetyError{
			code:    "CONFIRMATION_DECLINED",
			message: fmt.Sprintf("write operation for %q was not confirmed", commandPath),
		}
	}
	return nil
}

// lineReadResult carries either the entered text or its read error from the
// background reader back to readLine.
type lineReadResult struct {
	line string
	err  error
}

func readLine(ctx context.Context, in io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	reader := bufio.NewReader(in)
	if ctx.Done() == nil {
		return reader.ReadString('\n')
	}

	// Reading from a terminal can wait indefinitely for the user to press Enter.
	// Read in the background so this function can also notice cancellation, such
	// as Ctrl-C. The channel has room for one result, allowing the background read
	// to finish cleanly even if cancellation made this function return first.
	result := make(chan lineReadResult, 1)
	go func() {
		line, err := reader.ReadString('\n')
		result <- lineReadResult{line: line, err: err}
	}()

	// Return whichever happens first: command cancellation or completed input.
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case read := <-result:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return read.line, read.err
	}
}

// terminalStream reports whether a stream is connected to an interactive
// terminal instead of a pipe, redirected file, or agent-controlled buffer.
func terminalStream(stream any, isTerminal func(int) bool) bool {
	fd, ok := stream.(interface{ Fd() uintptr })
	return ok && isTerminal != nil && isTerminal(int(fd.Fd()))
}
