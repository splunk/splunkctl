# Adding Commands

## Decision: generated vs override

| Scenario | Approach |
|---|---|
| REST command definition changes and the compatible canonical XML is available | Update the XML and re-run codegen |
| Generation behavior changes for commands described by the XML | Update the parser, generator, or template and re-run codegen |
| An object needs hand-written or multi-step behavior (poll, aggregate, etc.) | Prefer a full-object override in `cmd/override/` |
| A new Splunk REST object is not represented in the available XML | Add a complete override file, or update the matching canonical XML and regenerate |
| An existing `index`, `user`, or `scheduler` overlay changes | Audit the generated parent, override verbs, explicit registration, and regeneration behavior together |

## Override models

The preferred model for new override work is a **full-object replacement**. A file named `cmd/override/<object>.go` makes codegen skip the same-name XML object. Codegen only writes files; it does not delete the previously emitted `cmd/generated/<object>.go`. Introducing a full replacement therefore requires intentionally deleting that stale emitted file and updating both generated and override registration as applicable.

The repository also has **legacy explicit partial overlays** in `cmd/register_overrides.go`:

- `index` adds an override verb to the checked-in generated parent;
- `user` replaces generated verbs and adds another verb; and
- `scheduler` adds an override verb to the checked-in generated parent.

These overlays depend on explicit `AddCommand` and `RemoveCommand` calls; codegen does not create the composition. Their same-name override files cause future generator runs to skip the generated parents, so a change to one of them must audit the parent, override verbs, registration, and artifact ownership together.

Do not introduce a new partial overlay casually. It requires an explicit design, behavior tests for registration and dispatch, a documented generator and artifact-set strategy, and updated contributor documentation. Prefer a full-object replacement when the override should own the object's command surface.

## Adding a full-object override command

Use this when the command cannot be expressed as a single REST call (for example, `search`, which submits a job, polls, and fetches results), or when a whole object needs hand-written behavior.

### Step 1: Create the file

```go
// cmd/override/myobject.go
package override

import (
    "github.com/spf13/cobra"
    "github.com/splunk/splunkctl/internal/client"
    "github.com/splunk/splunkctl/internal/cmdctx"
    "github.com/splunk/splunkctl/internal/output"
)

func newMyobjectCmd() *cobra.Command {
    objectCmd := &cobra.Command{
        Use:   "myobject",
        Short: "Manage myobject resources",
    }
    listCmd := &cobra.Command{
        Use:   "list",
        Short: "List all myobject resources",
        Args:  cobra.NoArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            c := cmdctx.ClientFrom(cmd)
            results, err := c.Do(cmd.Context(), client.Request{
                Method: "GET",
                Path:   "/myobject/",
            })
            if err != nil {
                return err
            }
            outFmt, _ := cmd.Flags().GetString("output")
            return output.Print(cmd.Context(), results, outFmt)
        },
    }
    objectCmd.AddCommand(listCmd)
    return objectCmd
}

var MyobjectCmd = newMyobjectCmd()
```

The package-local constructor gives production one registered instance while allowing same-package tests to build a fresh command tree without sharing mutated Cobra state. A cohesive package-local `RunE` handler is also acceptable when that is the clearer test boundary.

### Step 2: Register it

Add to `cmd/register_overrides.go`:

```go
rootCmd.AddCommand(
    override.SearchCmd,
    override.JobsCmd,
    // ...
    override.MyobjectCmd,  // add this
)
```

If this replaces a generated object, also remove its entry from `cmd/register_generated.go` and delete the old emitted object file. Confirm that no other registration or test still depends on that generated parent.

### Step 3: Write a test

Create `cmd/override/myobject_test.go` in package `override` so it can call the package-local constructor, and exercise the changed command or its responsible handler. A direct `client.Do` call against a test server is useful for REST transport and request semantics, but it does not prove that Cobra wires the command correctly and is insufficient by itself for a command change.

Use a fresh command from the package-local constructor, or a fresh test command around the cohesive handler. Use existing package test-server helpers where they fit; there is no repository-wide command-construction helper. Command-level coverage must, as applicable:

- invoke the command or handler with realistic positional arguments and flags;
- inject the test client and other dependencies through the command context;
- assert the resulting HTTP method, path, query, and form body;
- capture and assert structured output through the injected output writer, or configure and assert Cobra writers for an intentional raw, fixed-format, interactive, or progress contract; and
- cover relevant validation and error paths as well as success.

Do not reproduce the request directly in the test and treat that as command coverage. See [testing.md](testing.md) for the command-boundary pattern, package locations, state-isolation rules, and test-server helper details.

**If you add or modify a command, make sure to add a corresponding integration test for that command.** Unit tests cover request/response contracts and Cobra wiring against a mock server; integration tests confirm the command actually works against a real Splunk instance. See [Integration Tests](testing.md#integration-tests) in testing.md for the config/golden-file structure, a worked example, and the generate → test workflow.

### Step 4: Run focused and required checks

```bash
go test ./cmd/override/...
make fmt-check
go vet ./...
make test
make build
go run . myobject list --help
```

## Changing generated behavior

Within `cmd/generated/`, an object file carrying this header is an immutable emitted artifact:

```go
// Code generated by splunkctl codegen. DO NOT EDIT.
```

Do not edit those emitted object files directly. `cmd/generated/shared.go`, `_test.go` files, and `.gitkeep` do not carry the generated header and are hand-maintained.

Change generated behavior at the source that owns it:

1. Update the compatible canonical `splunkrc_cmds.xml` when the REST command definition is changing.
2. Update `codegen/parser.go`, `codegen/generator.go`, or `codegen/templates/command.go.tmpl` when generation behavior is changing.
3. Add or change a preferred full-object override when an object needs hand-written behavior. Remove the previously emitted object file and update registration because a generator skip does not delete stale output.

The existing `index`, `user`, and `scheduler` partial overlays are the explicit legacy exception described above. Changes to them must account for both the generated parent and override wiring; they are not a general model for mixing generated and overridden verbs.

Then follow [codegen.md](codegen.md) and commit the source and emitted object-file changes together. If the compatible external XML prerequisite is unavailable for a change that requires regeneration, stop and disclose that prerequisite. Do not patch emitted output or claim that regeneration succeeded.

## Rules for all commands

**Command shape:** Keep commands resource-first: `splunkctl <object> <verb>`. Do not introduce `splunkctl <verb> <object>` forms.

**Output:** Use `output.Print(cmd.Context(), results, outFmt)` for structured resource and REST results. Raw or fixed-format responses, interactive prompts, and progress displays may intentionally use Cobra writers or an established terminal-specific path; tests must cover that output contract. Do not add direct process-stream writes unless an established special path requires them.

**Client access:** Always call `cmdctx.ClientFrom(cmd)` (or the package-level `clientFrom(cmd)` shim in generated files). Never construct a new client inside a command.
Use the client's request methods, including `DoHTTP` for a caller-built request, so unsafe HTTP methods pass through the shared write check.

**Flag reads:** Read all flags inside `RunE`, not in package-level `init()`. This avoids global state issues in tests.

**Context propagation:** Pass `cmd.Context()` to `c.Do()` and to `output.Print()`. Don't use `context.Background()`.

**No-client commands:** Add `Annotations: map[string]string{"no_client": "true"}` for commands that don't need auth (version, schema, config, skill). Without this annotation, `PersistentPreRunE` will require `--host` and `--token`.

**Local writes:** Call `cmdctx.CheckWrite(cmd.Context())` immediately before a command's first filesystem or other local state change.
The invocation-scoped check prompts at most once, so a multi-step command must reuse the same command context throughout.

**Error messages:** Return `fmt.Errorf("...")` with lowercase first word (Go convention). Don't print errors yourself — let `cmd.Execute()` → `errorToCode()` handle them.

## Output flag wiring

The `--output`/`-o` flag is a persistent flag on the root command and is available to all commands. Inside `RunE`, read it with:

```go
outFmt, _ := cmd.Flags().GetString("output")
return output.Print(cmd.Context(), results, outFmt)
```

When `outFmt` is empty, `output.Print` auto-detects: table when stdout is a TTY, json otherwise.

## Conditional request parameters in override commands

When a command needs to vary request parameters based on argument shape, add a package-local helper and call it inside `RunE`. The `app install` override in `cmd/override/app.go` is an example: a local `isFilePath` function detects file-path arguments (by prefix `/`, `./` or extension `.tgz`, `.spl`, `.tar.gz`, `.tar`) and adds `filename=true` to the POST body when matched. Test the helper directly in a `*_unit_test.go` file in package `override`.

## Multi-step commands (like search)

Pattern from `cmd/override/search.go`:

```go
// Step 1: submit — use RawDo for operations that need raw response bytes
body, status, err := c.RawDo(cmd.Context(), "POST", "/search/jobs", form)

// Step 2: poll — loop with select{ctx.Done, time.After(2*time.Second)}
// Report progress to os.Stderr only when term.IsTerminal(int(os.Stderr.Fd()))

// Step 3: fetch and render
return output.Print(cmd.Context(), results, outFmt)
```

`RawDo` returns `([]byte, int, error)` — raw body, HTTP status, transport error. Check status >= 400 yourself.
