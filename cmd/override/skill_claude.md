---
name: splunkctl
description: Use for Splunk administration via CLI — searching, index/user/app management, cluster ops, KV Store, dashboards, saved searches. Provides reliable REST API interaction with built-in error handling and security.
---

# splunkctl — Splunk CLI Agent Skill

Use this skill when the user asks you to manage a Splunk instance: run searches, list/create/modify indexes, manage users, query cluster status, configure apps, or administer any Splunk resource.

## MCP vs splunkctl

If a Splunk MCP server is available, prefer it — it has lower overhead. Use splunkctl when:
- The MCP server doesn't support the operation you need
- The MCP server is unavailable
- You need the full REST API surface (run `splunkctl schema --groups` for discovery)

## Setup & Configuration

### Minimal setup (recommended for agents)

```bash
export SPLUNKCTL_HOST=https://your-splunk:8089
export SPLUNKCTL_TOKEN=your-api-token
splunkctl whoami --output json  # verify credentials and connectivity
```

**Configuration resolution chain** (first match wins): CLI flag → environment variable → `~/.splunkctl/config.yaml`

### Local development with self-signed certs

```bash
export SPLUNKCTL_HOST=https://localhost:8089
export SPLUNKCTL_TOKEN=your-token
splunkctl --insecure index list --output json
```

Or per-command: `splunkctl --insecure index list`

### Persistent config

```bash
splunkctl config set host https://your-splunk:8089
splunkctl config set token your-token
# Config stored in ~/.splunkctl/config.yaml (0600 permissions, never commit to git)
```

## Discovery Workflow

Start with discovery to avoid blind guessing:

```bash
# Step 1: available command groups
splunkctl schema --groups

# Step 2: explore a group (search, index, user, cluster, etc.)
splunkctl schema --group search --compact

# Step 3: inspect command parameters
splunkctl index add --describe        # lists all params for index creation
splunkctl user add --describe         # lists all params for user creation
```

## Common Patterns

### Search

```bash
# Basic search with JSON output (default max 100 results)
splunkctl search 'index=main error' --output json

# Increase result limit (max 50000 per command)
splunkctl search 'index=main status>=500' --maxout 1000

# Long-running search: detach and poll later (TTL: 1 hour)
splunkctl search 'index=main earliest=-7d' --detach
# Returns: {"job_id": "<sid>", "ttl": 3600}
# Later: splunkctl jobs show <sid>

# Stream search results without creating a persistent job
splunkctl search export 'index=main' --output json
```

### Index Management

```bash
splunkctl index list --output json
splunkctl index list --name myindex --output json
splunkctl index add --name newindex --param maxTotalDataSizeMB=10240
splunkctl index disable myindex
splunkctl index enable myindex
```

### User Management

```bash
splunkctl user list --output json
splunkctl user add jdoe --password '<new-user-password>' --role user
splunkctl user edit jdoe --password '<new-password>'
splunkctl user remove jdoe
```

### Apps

```bash
splunkctl app display --output json
splunkctl app enable myapp
splunkctl app disable myapp
splunkctl app install /path/to/app.tar.gz
```

### Saved Searches & Scheduled Alerts

```bash
splunkctl saved-search list --output json
splunkctl saved-search add "Alert Name" \
  --search 'index=main error | stats count' \
  --param cron_schedule="0 * * * *" \
  --param is_scheduled=1
splunkctl fired-alert list --output json
```

### HEC Event Send

```bash
# Send a single JSON event
splunkctl hec send --event '{"event":"data","source":"app"}' \
  --hec-token <hec-token> \
  --host https://splunk:8088

# Local dev with self-signed cert
splunkctl hec send --event '{"event":"test"}' \
  --hec-token <hec-token> \
  --host https://localhost:8088 \
  --insecure
```

### Dashboards

```bash
splunkctl dashboard list --output json
splunkctl dashboard show my_dashboard
splunkctl dashboard add my_dashboard --file dashboard.xml
```

### Knowledge Objects

```bash
# Lookups
splunkctl lookup list --output json
splunkctl lookup add my_lookup --filename my_lookup.csv

# Macros
splunkctl macro list --output json
splunkctl macro add my_macro --definition 'index=main sourcetype=syslog'

# Event types
splunkctl eventtype list --output json
splunkctl eventtype add web_errors --search 'index=main status>=400'

# Field extractions
splunkctl field-extraction list --output json

# Data models
splunkctl data-model list --output json
splunkctl data-model acceleration enable Authentication
```

### KV Store (app-scoped collections)

```bash
splunkctl kv collection list --app myapp --output json
splunkctl kv data put --app myapp --collection mycoll \
  --data '{"_key":"1","name":"Alice"}'
splunkctl kv data query --app myapp --collection mycoll --output json
```

### Cluster Operations

```bash
splunkctl cluster-peers list --output json
splunkctl shcluster-status show --output json
splunkctl health list --output json
```

## Output & Error Handling

**Output modes** (auto-detected in agent context):
- `--output json` — Always use for scripting/automation
- `--output table` — Human-readable (TTY only)
- `--output text` — Minimal output

**Exit codes**:
| Code | Meaning | Action |
|---|---|---|
| `0` | Success | — |
| `1` | General error / usage | Check command syntax |
| `3` | Not found (HTTP 404) | Resource doesn't exist or wrong namespace |
| `4` | Auth error (HTTP 401/403) | Verify token, check role permissions |
| `5` | Connection error | Verify host, network connectivity |

**Errors to stderr** as JSON: `{"error":"...","code":"NOT_FOUND"}`

## Global Flags

| Flag | Env var | Default | Notes |
|---|---|---|---|
| `--host` | `SPLUNKCTL_HOST` | required | Supports `host`, `host:port`, `https://host:port` |
| `--token` | `SPLUNKCTL_TOKEN` | required | Bearer token from Splunk UI or API |
| `--output` | — | auto | `json`, `table`, or `text` |
| `--insecure` | — | false | Skip TLS verification (local dev only) |
| `--param key=val` | — | — | Pass extra REST API parameters (repeatable) |
| `--describe` | — | false | Show available parameters for a resource (no mutation) |

## Troubleshooting

### TLS / Certificate Errors

```
x509: certificate signed by unknown authority
```

**Cause:** Splunk using self-signed or internal CA certificate.

**Solution:**
```bash
splunkctl --insecure index list
```

⚠️ **Security note:** `--insecure` disables all TLS verification. Use only for local development/testing.

### Connection / Timeout Errors

```
connection refused | i/o timeout | connection reset
```

**Common causes:**
- Wrong host/port (default: `8089` for REST management API, `8088` for HEC)
- Splunk service not running
- Network connectivity issue
- 30-second operation timeout (hard limit)

**Verify connection:**
```bash
splunkctl whoami --output json   # lightweight authenticated test
```

### Auth Errors

```
Exit 4: AUTH_ERROR (HTTP 401/403)
```

**Cause:** Invalid/expired token or insufficient permissions.

**Solution:**
```bash
# Regenerate token from Splunk UI (Settings → Tokens)
splunkctl config set token <new-token>

# Verify permissions (requires admin role)
splunkctl user list
```

### Rate Limiting (HTTP 429)

If you see `RATE_LIMITED` errors in scripts, the Splunk instance is throttling requests.

**Solution:** Add retry logic external to splunkctl, or space out commands:
```bash
for idx in {1..100}; do
  splunkctl index add myindex-$idx && sleep 0.5 || break
done
```

### Resource Not Found (Exit 3)

```
Exit 3: NOT_FOUND
```

**Cause:** Object doesn't exist, or searching wrong app/owner namespace.

**Debug:** List what exists:
```bash
splunkctl saved-search list --output json | jq '.[] | .name'
```

## Command Discovery

Full REST API surface available:

```bash
splunkctl schema --compact  # all commands as tree

# OR discover a specific object's parameters:
splunkctl index add --describe      # shows required/optional fields
splunkctl user add --describe       # shows user creation fields
splunkctl saved-search add --describe  # shows saved-search fields

# Use --param key=val for fields not exposed as flags:
splunkctl index add --name myindex --param maxTotalDataSizeMB=10240
```

## Skill Installation

Install this skill to make splunkctl available to Claude Code or Codex:

```bash
# Interactive (prompts for agent and scope)
splunkctl skill install --interactive

# Non-interactive
splunkctl skill install --agent claude --global
splunkctl skill install --agent codex --local
splunkctl skill install --agent both --global
```

**After installation:** Make sure splunkctl is on your PATH:
```bash
cp splunkctl /usr/local/bin/
# OR
export PATH="$PATH:$(pwd)"
```

## Production Checklist

Before deploying splunkctl in production/CI:

- [ ] **Credentials:** Use `SPLUNKCTL_TOKEN` env var, never hardcode tokens
- [ ] **TLS:** Only use `--insecure` in local dev; disable for production
- [ ] **Timeout awareness:** Long operations (bulk KV Store, large searches) may hit 30-second limit; use `--detach` for fire-and-forget
- [ ] **Output format:** Always use `--output json` for automation; avoid parsing text/table output
- [ ] **Error handling:** Check exit codes (0 = success, 3 = not found, 4 = auth, 5 = connection)
- [ ] **Rate limiting:** Implement external retry logic for parallel operations
- [ ] **Testing:** Run `splunkctl whoami --output json` to verify connectivity before bulk operations
- [ ] **Integration tests:** Write integration tests for any new/additional commands and run the validation pipeline (see [testing.md](../../docs/testing.md#integration-tests))
- [ ] **Privileges:** Splunk token needs appropriate admin/power roles for mutation commands
