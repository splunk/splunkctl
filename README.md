# splunkctl

[![Go Version](https://img.shields.io/github/go-mod/go-version/splunk/splunkctl)](go.mod)
[![License](https://img.shields.io/github/license/splunk/splunkctl)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/splunk/splunkctl)](https://github.com/splunk/splunkctl/releases)

`splunkctl` is a command-line interface for operating Splunk over its REST API. It is built for callers that aren't people — AI agents, CI jobs, scheduled scripts — and works the same when a person runs it.

Operating Splunk is mostly investigation: running searches, reading configuration, checking cluster state, working out why something behaves the way it does. A smaller part of it changes something. `splunkctl` treats those two differently. Investigation runs unattended. Change stops for a person.

> **splunkctl is experimental.** It isn't covered by existing Splunk support contracts. Review [Support](#support) before using it.

## What it covers

`splunkctl` manages Splunk deployments across searches, indexes, users, apps, saved searches, lookups, KV Store, HTTP Event Collector, and cluster internals, including indexer and search-head operations. Its commands are generated from Splunk's REST API definitions rather than maintained as a limited, hand-picked set, so new API endpoints can become available as commands. Since everything runs through REST, nothing needs to be installed on the Splunk server, and the CLI works consistently from a laptop, CI pipeline, or container against any Splunk instance your credentials can access.

## Installation 📦

### Manual Download
Download a ready-to-use binary from the [GitHub Releases](https://github.com/splunk/splunkctl/releases) page. Go is not needed to use a release binary.

#### Linux or macOS:
```bash
# Make a dir and untar the folder contents
mkdir splunkctl && tar -xzf splunkctl_<version>_<platform>.tar.gz -C splunkctl
cd splunkctl 
# Try running the binary now
./splunkctl version
```
**Note**: See [Troubleshooting Section](#troubleshooting-and-faqs-) for security issues with the downloaded binary on macOS 

#### Windows
On Windows, download and extract the matching .zip file, then put splunkctl.exe in a folder on your PATH.

### Build from source
``` bash 
git clone https://github.com/splunk/splunkctl
cd splunkctl
go build -o splunkctl .
```
Once installation finishes, point it at a Splunk instance (see [Configure SplunkCTL](#configure-splunkctl) below).

## Configure SplunkCTL ⚙️

`splunkctl` requires a Splunk host and an API bearer token.

**Step 1 — Get a token**
In Splunk Web: **Settings → Tokens → New Token**. Copy the token value.

Or via Splunk CLI on the Splunk host:
```bash
splunk create-authtokens -name mytoken -user admin
```

**Step 2 — Configure splunkctl** The easiest way is to save to the config file once:

```bash
splunkctl config set host https://your-splunk-host:8089
splunkctl config set token <your-token>
```

This writes `~/.splunkctl/config.yaml`. From then on you don't need to pass flags.

Alternatively, use flags or env vars per-invocation:

| Method | Host | Token |
|---|---|---|
| Flag | `--host https://splunk:8089` | `--token <token>` |
| Env var | `SPLUNKCTL_HOST=https://splunk:8089` | `SPLUNKCTL_TOKEN=<token>` |
| Config file | `splunkctl config set host ...` | `splunkctl config set token ...` |

Priority: flag > env var > config file.

**Step 3 — Verify**

```bash
splunkctl index list
```

If you see a table of indexes, you're connected.

## Usage 💻

Commands follow the pattern `splunkctl <object> <verb>`:

```bash
# List indexes
splunkctl index list
# Add a user
splunkctl user add jdoe --role admin --password changeme
# Search
splunkctl search 'index=main error' --maxout 100
# Async search (return job ID immediately)
splunkctl search 'index=main error' --detach
# Retrieve job results
splunkctl jobs show <sid>
# Show cluster config
splunkctl cluster-config list
# Create a workload rule (routes matching searches into a resource pool)
splunkctl workload-rule add my_rule --predicate "app=search" --workload_pool high_perf --schedule always_on
# Put a search head cluster into maintenance mode
splunkctl shcluster-maintenance-mode enable
```

## Write safety

- Use `--read-only` to block local changes and Splunk API requests that could modify data.
- Use `--yes` to skip the confirmation prompt for that command.
- Without either flag, interactive terminals ask for confirmation before the first write. Non-interactive callers receive `CONFIRMATION_REQUIRED`.
- If both flags are used, `--read-only` takes priority.
- The check is intentionally strict. Even a POST endpoint that only reports status is blocked.

## Discover available commands 🔍

Commands are organized into groups: **core** (shown in top-level `--help`),
**cluster** (shown under "Clustering Commands:"), and a larger set of
advanced/low-level commands that are callable directly but hidden from
`--help` — use `schema` or `<object> --help` to find them.

```bash
# Full command tree as JSON (useful for LLM agents)
splunkctl schema
# Compact view (commands + flags, no long descriptions)
splunkctl schema --compact
# List command groups
splunkctl schema --groups
# Filter by group
splunkctl schema --group index
# Top-level help
splunkctl --help
# Help for a specific object
splunkctl index --help
splunkctl user --help
```

The `splunkctl` schema entry lists the global flags inherited by every command.

## Output formats 📄

Structured resource commands support `--output` (short: `-o`). Commands with an explicit raw, fixed-format, interactive, or progress-output contract follow their command help instead.

| Value | Description |
|---|---|
| `table` | Human-readable columns (default on terminal) |
| `json` | Machine-readable JSON (default when piped/CI/agent) |
| `text` | One normalized map row per line |

```bash
splunkctl index list --output json
splunkctl index list --output table
```

## Working with AI Agent 🤖

### Skills: Using `splunkctl` CLI with AI Agents
The splunkctl skill helps Claude Code and Codex choose the right CLI command for a plain-language Splunk task and summarize the result.
Install the binary and skill first, following the [skill installation and setup instructions](cmd/override/skill_claude.md#skill-installation).
Then load it with `/skill splunkctl` in Claude Code or `$splunkctl` in Codex.

### Discovery and Usage for AI Agents
splunkctl is designed to be progressively discoverable by AI agents:

**Step 1 — Bootstrap:** Run `splunkctl schema --compact` to get the full command tree as JSON. This tells you every available command and its flags in one call (~50KB).

**Step 2 — Explore:** Run `splunkctl <object> --help` to see all operations on a resource.

**Step 3 — Execute:** For structured resource commands, use `--output json` (or rely on automatic JSON mode when stdout is not a TTY) for machine-parseable output. For raw, fixed-format, interactive, or progress commands, follow the command's help.

**Write approval:** When a command returns `CONFIRMATION_REQUIRED`, show the user a sanitized summary of the intended operation and its non-sensitive target, ask for approval, and rerun it with `--yes` only after the user agrees.
Never expose credentials or sensitive payloads while requesting approval.
Never remove `--read-only` or pipe confirmation input into the command.

**Environment variables:** All connection flags have `SPLUNKCTL_*` env var equivalents. Agents can configure splunkctl without modifying files.

**JSON output:** Structured resource commands automatically use JSON when stdout is not a terminal (for example, when an agent pipes the output). Their callers do not need to add `--output json`; commands with another explicit output contract follow their command help.

**Error format:** Errors always go to stderr as structured JSON containing only `error` and `code`:
```json
{"error": "index not found", "code": "NOT_FOUND"}
```

## Running unattended

An agent pointed at a stack can search, read index and app configuration, walk cluster state, and pull saved searches without clearing each step with someone. A diagnosis runs to the end.

A command that would change something stops before it does:

```bash
$ splunkctl index disable prod_logs
{"error":"write operation for \"splunkctl index disable\" requires confirmation; rerun with --yes after user approval","code":"CONFIRMATION_REQUIRED"}
```

The agent comes back with the change it intends and the target it resolved. Rerunning with `--yes` performs it.

How that behaves:

- Approval is given per run, on the command line. No environment variable or configuration setting substitutes for it.
- `--read-only` blocks change outright and takes precedence over approval.
- `splunkctl` makes no permission decisions of its own. A caller reaches what the token's owner can reach, and Splunk enforces and records it as that person.

## Repository structure 🗂️

A quick map for navigating the codebase:

```
splunkctl/
├── cmd/
│   ├── generated/       # Auto-generated commands — do not hand-edit (see docs/codegen.md)
│   ├── override/        # Hand-written fixes/replacements for specific generated commands
│   └── *.go             # CLI entrypoint, command registration, schema/config commands
├── codegen/             # The generator itself — parses the REST API definition, emits cmd/generated/
├── internal/            # Shared client, config, and output-formatting code
├── docs/                # Architecture, testing, codegen, and contribution docs
├── cicd/                # CI pipeline definitions and Orca deployment tooling for integration tests
└── tests/integration/   # Golden-file integration tests, one config per command
```

Fixing a broken command? Start in `cmd/generated/` and `cmd/override/` —
see [Code generator](docs/codegen.md#override-escape-hatch) for the pattern.

## Development 🛠️

If a generated command has a bug (wrong parameter name, wrong HTTP method,
missing required argument), fix it via a partial override in `cmd/override/`
rather than hand-editing `cmd/generated/`, which is regenerated from an
external XML source. See the "Override escape hatch" section in
[Code generator](docs/codegen.md#override-escape-hatch) and existing
examples like `cmd/override/app.go`.

```bash
make build
./bin/splunkctl --help
```

Run the same format, vet, test, and build checks required by CI before submitting a change:

```bash
make fmt-check
go vet ./...
make test
make build
```

## Docs and References 📚

| Doc | What it covers | When to use it |
|---|---|---|
| [Architecture](docs/architecture.md) | Package layout and the request lifecycle from CLI invocation to REST call. | You're trying to understand how a command flows through the codebase, or where a new piece of logic belongs. |
| [Testing](docs/testing.md) | [Unit test](docs/testing.md#unit-tests) patterns and boundaries, plus [integration test](docs/testing.md#integration-tests) structure and workflow. | You're writing or debugging a test, or need to know whether a change needs unit coverage, integration coverage, or both. |
| [Adding commands](docs/adding-commands.md) | The generated-vs-override decision, the override escape hatch, and the rules every command must follow. | You're adding a new command or fixing a bug in an existing one. |
| [Code generator](docs/codegen.md) | How `cmd/generated/` is produced from the Splunk REST API definition, and how to regenerate it. | You need to understand or change what the generator produces, or you're fixing a bug that traces back to codegen behavior. |
| [Design decisions](docs/decisions.md) | Non-obvious architectural choices and the reasoning behind them. | Something in the codebase looks unusual and you want to know if it's intentional before changing it. |

## Troubleshooting and FAQs 🩹

#### 1. macOS security and quarantine errors

If macOS shows an error such as “cannot be opened because the developer cannot be verified,” try either of the following:

- Remove the quarantine attribute for the binary obtained from the official GitHub release page:

  ```bash
  xattr -d com.apple.quarantine ./splunkctl
  ```

- Go to System Settings -> Privacy & Security -> scroll down, and click "Allow Anyway" next to the blocked splunkctl message, then try running it again.


#### 2. Token stopped working after a secret rotation? 
Rotating a cluster's
`splunk.secret` invalidates existing bearer tokens across the cluster —
generate a new one (see [Configure SplunkCTL](#configure-splunkctl)).

### Compatibility ✅

Tested in CI against Splunk 10.4.0 and 10.4.2.

Supported platforms:

| Platform | Archive |
|---|---|
| Linux x86_64 | `linux_amd64` |
| Linux ARM64 | `linux_arm64` |
| macOS Intel | `darwin_amd64` |
| macOS Apple Silicon | `darwin_arm64` |
| Windows x86_64 | `windows_amd64` |


### Support

The commands in this repository are experimental and are not covered by
existing Splunk support contracts.

GitHub pull requests are not accepted for this repository.

For help with supported Splunk products, see [Working with Splunk Support](https://www.splunk.com/support).
Splunk Support cases do not provide support for these experimental commands.

Questions about splunkctl itself can be directed to the
[`splunkctl-support`](https://github.com/orgs/splunk/teams/splunkctl-support) team.

## License 📜

Copyright 2026 Splunk Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
