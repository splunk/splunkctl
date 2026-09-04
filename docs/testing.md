# Testing

## Table of contents

- [Unit Tests](#unit-tests)
  - [Running unit tests](#running-unit-tests)
  - [Test structure](#test-structure)
  - [Choose the boundary that changed](#choose-the-boundary-that-changed)
  - [Existing TLS test-server helper](#existing-tls-test-server-helper)
  - [Request-contract pattern](#request-contract-pattern)
  - [Command-boundary pattern](#command-boundary-pattern)
  - [File naming convention](#file-naming-convention)
  - [Writing unit tests](#writing-unit-tests)
  - [Writing config tests](#writing-config-tests)
  - [Testing codegen](#testing-codegen)
- [Integration Tests](#integration-tests)
  - [Overview](#overview)
  - [Structure](#structure)
  - [Example: the `app` integration test](#example-the-app-integration-test)
  - [Writing a new integration test](#writing-a-new-integration-test)
  - [Step 1: Generate mode](#step-1-generate-mode)
  - [Step 2: Test mode](#step-2-test-mode)
  - [Running locally](#running-locally)
  - [Running against a local Splunk instance](#running-against-a-local-splunk-instance)
  - [CI behavior](#ci-behavior)
- [Override registration testing](#override-registration-testing)
- [Work in progress](#work-in-progress)

# Unit Tests

## Running unit tests

```bash
# all tests
go test ./...

# verbose, single package
go test ./internal/client/... -v
go test ./cmd/... -v
go test ./cmd/generated/... -v
go test ./cmd/override/... -v

# single test
go test ./cmd/... -v -run TestSmoke_IndexListAndRenderJSON

# codegen tests
go test ./codegen/... -v
```

## Test structure

| Package | Files | What it covers |
|---|---|---|
| `cmd` | `smoke_test.go` | Existing cross-package client and output contracts; these tests do not dispatch Cobra commands |
| `cmd` | `schema_test.go` | Schema command output structure |
| `cmd` | `config_cmd_test.go` | Config command flag/env resolution |
| `cmd` | `version_test.go` | Version command defaults |
| `cmd` | `testhelper_test.go` | Shared `newTLSTestServer` helper |
| `cmd` | `register_overrides_test.go` | Confirms specific override subcommands are present on their generated parent (presence/lookup only, not full dispatch) |
| `cmd/generated` | `<command>_test.go` | Existing generated request contracts: HTTP method, path, and params; most do not invoke the generated Cobra handler |
| `cmd/generated` | `testhelper_test.go` | Shared `newTLSTestServer` helper |
| `cmd/override` | `<command>_test.go` | Existing override request and response contracts; most do not invoke the override Cobra handler |
| `cmd/override` | `testhelper_test.go` | Shared `newTLSTestServer` helper |
| `codegen` | `parser_test.go` | XML parsing: URIs, args, methods, synonyms |
| `internal/client` | `client_test.go` | Bearer token injection, URL building, response parsing |
| `internal/config` | `config_test.go` | Config resolution priority (flag > env > file) |
| `internal/output` | `output_test.go` | JSON/table/text rendering, name-column ordering |

## Choose the boundary that changed

A direct `client.Do` or `client.RawDo` call is request-contract coverage. It can verify transport behavior, method, path, parameters, authentication, and response parsing. It does not execute Cobra argument validation, flag parsing, inherited flags, context dependency lookup, `RunE`, command registration, or output wiring.

A Cobra command change also needs command- or handler-boundary coverage. Exercise realistic arguments and flags, inject dependencies through the command context, capture the applicable output path, and assert both the outgoing request and returned output or error. A registration change needs behavior coverage in package `cmd`, such as lookup and dispatch through a fresh root or isolated root command; successful compilation alone does not prove that the intended command wins or is reachable.

## Existing TLS test-server helper

The `cmd`, `cmd/generated`, and `cmd/override` test packages each define their own package-local `newTLSTestServer` helper:

```go
func newTLSTestServer(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
    t.Helper()
    srv := httptest.NewTLSServer(handler)
    t.Cleanup(srv.Close)
    cfg := &config.Config{Host: srv.URL, Token: "test-token", Insecure: true}
    return client.New(cfg), srv
}
```

This helper returns a configured client and in-process HTTPS server. It does not construct, register, or execute a Cobra command.

The existing `cmd/override/testhelper_test.go` uses package `override_test`, so a same-package `override` test cannot call that helper. Such a test must create its server and client inline, as below, or define a separate helper in package `override`.

## Request-contract pattern

The following existing style is appropriate when the client request contract is the layer under test:

```go
func TestIndex_List(t *testing.T) {
    c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            t.Errorf("expected GET, got %s", r.Method)
        }
        if r.URL.Path != "/services/data/indexes/" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "entry": []any{
                map[string]any{"name": "main", "content": map[string]any{"maxDataSizeMB": 500}},
            },
        })
    }))

    results, err := c.Do(context.Background(), client.Request{
        Method: "GET",
        Path:   "/data/indexes/",
    })
    if err != nil {
        t.Fatal(err)
    }
    if len(results) != 1 {
        t.Fatalf("expected 1 result, got %d", len(results))
    }
    if results[0]["name"] != "main" {
        t.Errorf("expected main, got %v", results[0]["name"])
    }
}
```

This example verifies `client.Do`, the request path, and response normalization. It is not sufficient by itself for a change to `IndexListCmd` or its registration.

Useful low-level request details:

- The client prepends `/services` to all paths — handlers must assert `/services/data/...`
- Use `r.URL.RawPath` (not `r.URL.Path`) when the ID contains slashes (e.g. monitor input `/var/log/`) because Go's HTTP server URL-decodes `%2F` in `r.URL.Path`
- GET and DELETE: params go into query string. POST: params go into form body (`r.ParseForm()` + `r.FormValue(...)`)
- ID is URL-path-escaped and appended to the trimmed path: `/data/indexes/` + `myindex` → `/services/data/indexes/myindex`

## Command-boundary pattern

New command code should expose either a package-local constructor that returns a fresh command tree or a cohesive package-local handler that a fresh test command can call. The constructor is part of the new command implementation; no repository-wide command-construction helper currently exists. Same-package tests can invoke an unexported constructor directly.

For a constructor such as `newMyobjectCmd()` from [adding-commands.md](adding-commands.md), build the root dependencies that the command actually inherits and execute realistic input:

The example uses standard-library packages `bytes`, `context`, `encoding/json`, `net/http`, `net/http/httptest`, `strings`, and `testing`, plus the repository's Cobra, client, command-context, config, and output packages.

```go
func TestMyobjectListCommand(t *testing.T) {
    srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            t.Errorf("expected GET, got %s", r.Method)
        }
        if r.URL.Path != "/services/myobject/" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(map[string]any{
            "entry": []any{
                map[string]any{"name": "item1", "content": map[string]any{}},
            },
        }); err != nil {
            t.Errorf("encode response: %v", err)
        }
    }))
    t.Cleanup(srv.Close)
    c := client.New(&config.Config{
        Host: srv.URL, Token: "test-token", Insecure: true,
    })

    var structured, stdout, stderr bytes.Buffer
    ctx := cmdctx.WithClient(context.Background(), c)
    ctx = output.WithWriter(ctx, &structured)

    root := &cobra.Command{Use: "splunkctl"}
    root.PersistentFlags().String("output", "", "")
    root.AddCommand(newMyobjectCmd())
    root.SetContext(ctx)
    root.SetOut(&stdout)
    root.SetErr(&stderr)
    root.SetArgs([]string{
        "myobject", "list",
        "--output", "table",
    })

    if err := root.Execute(); err != nil {
        t.Fatal(err)
    }
    got := structured.String()
    if !strings.Contains(got, "name") || !strings.Contains(got, "item1") {
        t.Fatalf("unexpected table output: %s", got)
    }
    if strings.Contains(got, `"name"`) {
        t.Fatalf("expected table output, got JSON: %s", got)
    }
    // Assert stdout/stderr instead when the command intentionally uses
    // Cobra writers.
}

func TestMyobjectListCommandRejectsArgs(t *testing.T) {
    root := &cobra.Command{Use: "splunkctl"}
    root.AddCommand(newMyobjectCmd())
    root.SetArgs([]string{"myobject", "list", "unexpected"})

    if err := root.Execute(); err == nil {
        t.Fatal("expected an error for an unexpected positional argument")
    }
}
```

Add the command-specific flags inside the fresh constructor, just as production does. If testing a cohesive handler instead, create a fresh `cobra.Command`, assign that handler to `RunE`, define every local or inherited flag the handler reads, inject the same context and writers, and execute it with `SetArgs`.

Package-global Cobra commands retain parsed flag values, args, context, writers, and parent relationships when mutated. Prefer fresh constructors. If an existing global tree must be exercised, keep the test in the owning package, save and restore every mutated field or rebuild the affected state with `t.Cleanup`, and do not run tests that mutate the same tree concurrently or with `t.Parallel()`.

## File naming convention

One contract test file per source command file. If the source file also contains internal helper or utility functions, add a second unit test file:

| Source file | Contract test file | Unit test file |
|---|---|---|
| `cmd/override/acl.go` | `cmd/override/acl_test.go` | `cmd/override/acl_unit_test.go` |
| `cmd/override/hec.go` | `cmd/override/hec_test.go` | (none — no utility functions) |
| `cmd/generated/index.go` | `cmd/generated/index_test.go` | (none) |
| `cmd/schema.go` | `cmd/schema_test.go` | (none) |

## Writing unit tests

Use `*_unit_test.go` files to directly test utility and helper functions inside a source file.
Keep `*_test.go` files for contract-level tests that verify command HTTP request/response behavior.

Example for `acl.go`:

```
acl.go               — source file with command + utility functions
acl_test.go          — contract test (uses package override_test + mock server)
acl_unit_test.go     — unit test (uses package override to access unexported functions)
```

The `testhelper_test.go` in each package provides the shared `newTLSTestServer` helper used by contract tests.

## Writing config tests

Config resolution tests mutate process environment and package-global Viper state. Reset Viper before the test, register a final reset, and use `t.Setenv` so Go restores each variable's original value (including a value that existed before the test):

```go
func TestResolve_EnvFallback(t *testing.T) {
    config.ResetForTesting()
    t.Cleanup(config.ResetForTesting)
    t.Setenv("SPLUNKCTL_HOST", "https://env-host:8089")
    t.Setenv("SPLUNKCTL_TOKEN", "env-token")

    cfg, err := config.Resolve("", "", false, "")
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Host != "https://env-host:8089" {
        t.Errorf("got %s", cfg.Host)
    }
}
```

`t.Setenv` registers environment restoration with test cleanup. Because `t.Cleanup(config.ResetForTesting)` is registered first, the later `t.Setenv` cleanups restore the environment before the final Viper reset runs. `ResetForTesting()` clears shared Viper state without reading `~/.splunkctl/config.yaml`, so both setup and teardown are hermetic.

Tests or subtests that mutate process environment or Viper state must not call `t.Parallel()` or otherwise run concurrently with tests that mutate the same state.

## Testing codegen

`codegen/parser_test.go` tests XML parsing. To test a generator behavior, write inline XML and run it through `BuildGeneratedFile`:

```go
func TestBuildGeneratedFile_PathTemplate(t *testing.T) {
    item := Item{
        Obj: "index",
        Common: Common{URI: "/data/indexes/"},
        Cmds: []Cmd{
            {Name: "enable", URI: "/data/indexes/{name}/enable/"},
        },
    }
    gf := BuildGeneratedFile(item)
    if !gf.NeedsStrings {
        t.Error("expected NeedsStrings for path template")
    }
    if !gf.Verbs[0].PathHasTemplate {
        t.Error("expected PathHasTemplate")
    }
}
```

# Integration Tests

## Overview

Integration tests run the real, compiled `splunkctl` binary against a **live Splunk instance** and compare its actual output against a committed "golden" JSON file, using exact `reflect.DeepEqual` matching. They catch things unit tests can't: real REST API behavior, real authentication, and real end-to-end command output — at the cost of needing a reachable Splunk instance to run.

If you add or change a command's behavior in a way that affects what it sends to or receives from Splunk, add or update an integration test for it, in addition to unit tests. See [Adding commands](adding-commands.md) for when this is expected.

## Structure

Integration tests live under `tests/integration/`, one directory per command, grouped by command category:

```
tests/integration/<group>/test_<command>/
  configs/<command>.yml   — test cases: name + splunkctl command
  output/<testname>.json  — golden output file, one per test case
```

- `<group>` is a category such as `core`.
- `configs/<command>.yml` lists one or more test cases for that command. Each case has a `testname` and the exact `splunkctl_cmd` to run.
- `output/<testname>.json` holds the expected result for that exact test case, keyed by `testname`.

Test names in the YAML file must use the command name as a prefix (e.g., `test_app_install_success`, not `install_success`). Each test name maps to exactly one golden file in `output/`.

## Example: the `app` integration test

`tests/integration/core/test_app/configs/app.yml` defines seven test cases for the `app` command:

```yaml
app:
- testname: test_app_fetch_not_found_error
  splunkctl_cmd: 'app display nonexistentapp123 --output json'
- testname: test_app_url_install_error
  splunkctl_cmd: 'app install https://example.com/test_app.tgz --output json --yes'
- testname: test_app_install_success
  splunkctl_cmd: "app install ${CI_TEST_APP_PATH:-$PWD/tests/resources/TestApp/test_app.tgz} --output json --yes"
- testname: test_app_display_multiply_success
  splunkctl_cmd: 'app display multiply --output json'
- testname: test_app_disable_success
  splunkctl_cmd: 'app disable multiply --output json --yes'
- testname: test_app_enable_success
  splunkctl_cmd: 'app enable multiply --output json --yes'
- testname: test_app_remove_success
  splunkctl_cmd: 'app remove multiply --output json --yes'
```

Each `testname` has a matching golden file in `tests/integration/core/test_app/output/`. For example, `test_app_display_multiply_success.json` holds the exact JSON the runner expects `app display multiply --output json` to produce:

```json
{
    "job_is_successful": true,
    "splunkctl_cmd": "app display multiply --output json",
    "msg_fatal": null,
    "msg_warns": null,
    "result_message": "Results set with 1 result(s)",
    "result_statistics": [
        {
            "name": "multiply",
            "...": "..."
        }
    ]
}
```

The successful lifecycle cases build on each other in order: install, display, disable, enable, then remove.
They share state on the same live Splunk instance, so keep this kind of ordering dependency in mind when adding cases to an existing config.

Also notice `test_app_install_success` uses `${CI_TEST_APP_PATH:-$PWD/tests/resources/TestApp/test_app.tgz}` instead of a hardcoded path — see [CI behavior](#ci-behavior) for why.

## Writing a new integration test

1. Pick (or create) the command's config file: `tests/integration/<group>/test_<command>/configs/<command>.yml`.
2. Add a new entry with a `testname` prefixed with the command name, and the exact `splunkctl_cmd` to run (without `./splunkctl` or `--insecure` — the runner adds both automatically).
   Add `--yes` when the fixture performs a local state change or sends a guarded HTTP method, including a status endpoint implemented with POST.
3. Run the config in **generate mode** (see below) to produce the golden file from a real run.
4. Inspect the generated golden file before committing it — confirm it reflects the behavior you actually intend, not just whatever the live instance happened to return.
5. Commit the config change and the new golden file together.

## Step 1: Generate mode

Generate mode runs your configured command against a live Splunk instance and **writes its actual output as the new golden file**, overwriting whatever was there before:

```bash
make generate-integration INTEGRATION_CMD=app
```

Use this the first time you add a test case, and any time a command's legitimate output changes and the golden file needs to catch up. Only regenerate against a clean Splunk state — for example, a pre-existing installed app will produce a `409` response that becomes the new (wrong) baseline.

## Step 2: Test mode

Test mode is the default (`INTEGRATION_MODE=test`, or simply omitting `INTEGRATION_MODE`). It runs the same configured command and **compares** the actual output against the existing golden file using exact `reflect.DeepEqual` matching, reporting pass/fail with a structured diff on mismatch:

```bash
make test-integration INTEGRATION_CMD=app
```

`INTEGRATION_CMD` selects a config filename without `.yml`; omit it to run every integration test.

## Running locally

First build the CLI and integration-test runner, then provide a host and bearer token using the environment or the normal `splunkctl` configuration:

```bash
make build
export SPLUNKCTL_HOST=https://your-splunk-host:8089
export SPLUNKCTL_TOKEN=<your-token>
make test-integration INTEGRATION_CMD=server_health
```

`SPLUNKCTL_HOST` and `SPLUNKCTL_TOKEN` must be configured (via `~/.splunkctl/config.yaml` or env vars) pointing at a live Splunk instance. Add `--insecure` to the config if TLS verification should be skipped.

## Running against a local Splunk instance

1. Start Splunk locally and confirm that its management API is available (normally at `https://127.0.0.1:8089`).

2. Open [`cicd/tools/orca/deployment/local-deployment.yml`](../cicd/tools/orca/deployment/local-deployment.yml) and replace the placeholder admin credentials with the credentials for that Splunk instance. Change `host` or `port` too if the management API is listening elsewhere:

   ```yaml
   so1:
     host: 127.0.0.1
     splunk:
       port: 8089
       username: admin
       password: 'your-local-admin-password'
     access:
       type: local
   ```

   Keep `access.type` set to `local`. The account must be allowed to enable token authentication and create authentication tokens. Quote passwords containing YAML punctuation, and do not commit real credentials; restore the placeholder after testing.

3. Run one config or omit `INTEGRATION_CMD` to run all configs:

   ```bash
   make test-integration-local INTEGRATION_CMD=server_health
   ```

The local target builds both binaries, verifies the credentials, enables token authentication, and creates a one-hour ephemeral bearer token. It supplies the generated host and token only to the integration-test process and removes the temporary configuration afterward. It does not stop or delete the local Splunk instance.

To update golden files against the local instance, set generate mode explicitly:

```bash
make test-integration-local INTEGRATION_MODE=generate INTEGRATION_CMD=server_health
```

To keep credentials outside the tracked deployment file, copy it elsewhere and pass the copy:

```bash
make test-integration-local \
  LOCAL_DEPLOYMENT_FILE=/path/to/local-deployment.yml \
  INTEGRATION_CMD=server_health
```

The integration runner uses enhanced output by default, including per-test timing, summaries, and structured JSON differences on failures. Its `--output-format` flag also accepts `simple` and `json` when invoking `bin/integration-runner` directly.

## CI behavior

CI (ORCA) runs integration tests the same way as locally, but some tests need a file staged on the Splunk server's own filesystem first — the app integration tests are the current example.

| Context | Path |
|---|---|
| CI (ORCA) | `/opt/splunk/var/run/splunk/test_app.tgz` (set via `CI_TEST_APP_PATH`) |
| Local | `$PWD/tests/resources/TestApp/test_app.tgz` (shell fallback) |

The config YAML uses `${CI_TEST_APP_PATH:-$PWD/tests/resources/TestApp/test_app.tgz}` so both environments work without modification — see the `test_app_install_success` case in [the `app` example above](#example-the-app-integration-test). In CI, `orca_service.py` stages the archive onto the ORCA container via `scp` during the deployment phase so the file is present on the Splunk server before tests run.

## GitHub publication tests

The publication helper has hermetic tests that create temporary local Git
repositories. They cover committed-blob packaging, exclusion enforcement,
preservation of unmanaged GitHub files, publication to a non-protected branch,
repeat no-op behavior, and artifact-tamper rejection:

```bash
go test ./cicd/tools/github-publication
```

In CI, `validate:github` also runs `go test`, `go vet`, `go build`, root help,
version JSON, and schema generation from the packaged public candidate. The job
then compares that candidate with GitHub `main` and stores the exact diff and
checksummed validation artifact consumed by the manual publish job.

## GitHub public projection

[github.com/splunk/splunkctl](https://github.com/splunk/splunkctl). GitHub-only
paths outside the declared managed roots, such as repository settings under
`.github/`, are preserved.

The publication flow has four jobs:

1. `package:github` builds the allowlisted candidate from the committed GitLab
   revision and scans it for secrets.
2. `validate:github` tests that candidate and creates an exact, checksummed diff
   against GitHub `main`.
3. `update:github` is a manual job available only for protected GitLab tags. It
   rejects an altered artifact or a GitLab/GitHub branch that moved after
   validation, pushes the tag-specific `gitlab-release/<tag>` PR branch, and
   opens or reuses a PR into GitHub `main`. GitHub `main` is never pushed
   directly.
4. `release:github` is a manual job for the same protected tag. It consumes the
   GitLab-built archives and the publication result from `update:github`.

# Override registration testing

Several override commands added for recent clustering-command fixes have command-boundary tests that build a fresh Cobra tree around the override's own constructor, dispatch through `root.Execute()` with realistic args, and assert the actual outgoing HTTP request:

| Command | Test file | Notes |
|---|---|---|
| `workload-rule add` | `cmd/override/workload_rule_add_test.go` | Asserts the request form contains `name`, not the old `rule_name`. |
| `workload-pool add` | `cmd/override/workload_pool_add_test.go` | Asserts the request form contains `name`, not the old `pool_name`. |
| `shcluster-maintenance-mode enable`/`disable` | `cmd/override/shcluster_maintenance_mode_param_test.go` | Asserts `mode=1`/`mode=0`, plus a regression test proving a conflicting `--param mode=...` cannot override the fixed subcommand semantics. |
| `maintenance-mode enable`/`disable` | `cmd/override/maintenance_mode_param_test.go` | Same coverage as the `shcluster-maintenance-mode` sibling above. |

Each of these tests was verified to actually catch a regression: the fix was temporarily reverted, the corresponding test was confirmed to fail, then the fix was restored.

This coverage proves the override's own request-building logic is correct. It does not yet exercise the full `register_overrides.go` → `rootCmd` registration path (confirming the generated parent's verb was actually removed and replaced at the real root level, as opposed to a fresh isolated tree) — see [Choose the boundary that changed](#choose-the-boundary-that-changed) for that distinction.

# Work in progress

- The `skill` command's successful install/list/remove lifecycle still needs full file-I/O coverage; current temporary-home tests cover mutation denial.
- Cluster/SHC request shapes have mock-server coverage, and a growing number of specific override commands now have full command-boundary/dispatch coverage (e.g. `shcluster-maintenance-mode`, `maintenance-mode`), but most `cmd/generated` cluster/SHC Cobra handlers still lack dispatch coverage.
