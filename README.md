# splunkctl

[![pipeline status](https://cd.splunkdev.com/ai/splunkctl/badges/main/pipeline.svg)](https://cd.splunkdev.com/ai/splunkctl/-/commits/main)

Welcome to `splunkctl`, a lightweight, agent-native CLI for managing Splunk through its REST API. It covers the full Splunk command surface — indexes, users, search, clustering, and more — with commands generated directly from the REST API definition, plus hand-written overrides where an operation needs multi-step logic (search, KV store data, etc.).

## Quickstart 🚀

Get up and running in seconds:

```bash
# Option A: run without installing
go run . --help
go run . version

# Option B: build a binary
go build -o splunkctl .
./splunkctl --help
```

Then point it at a Splunk instance (see [Configuration](#configuration) below).

## Install (standalone) 📦

```bash
go install github.com/splunk/splunkctl@latest
```

Or build from source:

```bash
git clone https://github.com/splunk/splunkctl
cd splunkctl
go build -o splunkctl .
```

## Configuration ⚙️

splunkctl requires a Splunk host and an API bearer token.

**Step 1 — Get a token**

In Splunk Web: **Settings → Tokens → New Token**. Copy the token value.

Or via Splunk CLI on the Splunk host:
```bash
splunk create-authtokens -name mytoken -user admin
```

**Step 2 — Configure splunkctl**

The easiest way is to save to the config file once:

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

## For AI Agents 🤖

### Skills: Using `splunkctl` CLI with AI Agents
The splunkctl skill helps Claude Code and Codex choose the right CLI command for a plain-language Splunk task and summarize the result.
Install the binary and skill first, following the [skill installation and setup instructions](cmd/override/skill_claude.md#skill-installation).
Then load it with `/skill splunkctl` in Claude Code or `$splunkctl` in Codex.

### Discovery and Usage for AI Agents
splunkctl is designed to be progressively discoverable by AI agents:

**Step 1 — Bootstrap:** Run `splunkctl schema --compact` to get the full command tree as JSON. This tells you every available command and its flags in one call (~50KB).

**Step 2 — Explore:** Run `splunkctl <object> --help` to see all operations on a resource.

**Step 3 — Execute:** For structured resource commands, use `--output json` (or rely on automatic JSON mode when stdout is not a TTY) for machine-parseable output. For raw, fixed-format, interactive, or progress commands, follow the command's help.

**Environment variables:** All connection flags have `SPLUNKCTL_*` env var equivalents. Agents can configure splunkctl without modifying files.

**JSON output:** Structured resource commands automatically use JSON when stdout is not a terminal (for example, when an agent pipes the output). Their callers do not need to add `--output json`; commands with another explicit output contract follow their command help.

**Error format:** Errors always go to stderr as structured JSON containing only `error` and `code`:
```json
{"error": "index not found", "code": "NOT_FOUND"}
```

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

**Token stopped working after a secret rotation?** Rotating a cluster's
`splunk.secret` invalidates existing bearer tokens across the cluster —
generate a new one (see [Configuration](#configuration)).

### Compatibility ✅

Tested in CI against Splunk 10.4.0 and 10.4.2.

## Contribution 🥰

Found a bug or have an idea for improving splunkctl? Open a branch, make
focused commits, and make sure the required checks pass (see
[Development](#development)) before submitting your change for review. See
[Adding commands](docs/adding-commands.md) if you're adding or fixing a
command.

Maintainers: [@gmeghan14](https://github.com/gmeghan14) and [@shruti148](https://github.com/shruti148).

### Support

Support channels for splunkctl are still being finalized. In the meantime,
reach out to a maintainer directly (see [Contribution](#contribution) above).

## License 📜

Copyright 2026 Splunk Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
