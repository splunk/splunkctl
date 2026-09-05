---
name: splunkctl
description: "Use for Splunk REST administration with the splunkctl CLI: searches, jobs, indexes, users, roles, apps, cluster status, knowledge objects, HEC, and config-backed management tasks."
allowed-tools:
  - Bash
  - Read
---

# splunkctl

Use this skill when the user asks to interact with Splunk through `splunkctl`.

## Before Running Commands

- Prefer a configured Splunk MCP tool only when it is already available and exposes the requested operation. Do not troubleshoot MCP configuration unless an MCP tool was actually attempted and failed.
- If `splunkctl` is not on `PATH` in this repository, run the same command as `go run . ...` from the repo root.
- Use only host and bearer-token auth: `--host`/`SPLUNKCTL_HOST` and `--token`/`SPLUNKCTL_TOKEN`. Do not use username/password login for management auth. Password flags are only for user or credential management commands.
- Authentication resolves as flags, then environment variables, then `~/.splunkctl/config.yaml`; prefer flags or environment variables for agents.
- Keep tokens and passwords out of chat output, command echoes, logs, and final answers.
- `splunkctl version --output json` verifies the CLI build, not the Splunk connection. For a live read-only check, use `splunkctl whoami --output json` or the requested list/show command.
- For local dev Splunk, use `https://localhost:8089`. If TLS fails because of a self-signed certificate, retry with `--insecure`.
- In sandboxed agent environments, loopback/network calls can fail with `operation not permitted`. If the command is necessary and in scope, request elevated execution and rerun the same command.

## Write Safety

- Use `--read-only` when the requested workflow must not change local or Splunk state.
- If a command returns `CONFIRMATION_REQUIRED`, give the user a sanitized operation summary and non-sensitive target, then rerun it with `--yes` only after approval.
- Never expose credentials or sensitive payloads while requesting approval.
- Treat `READ_ONLY` as a hard stop and never remove the flag or pipe confirmation input into the command.
- Approval applies only to the operation the user approved.

## Discovery

Do cheap discovery before choosing a command:

```bash
splunkctl schema --groups
splunkctl schema --group search --compact
splunkctl schema --group index --compact
splunkctl <object> --help
```

Useful group mapping:

- `auth`: `user` and `role`
- `apps`: `app`
- `search`: `search`, `jobs`, `saved-search`, `fired-alert`
- `index`: `index`, `indexer-discovery`, `indexing-ready`
- `server`: server info, messages, and health helpers

Run `splunkctl schema --compact` only when the group is unclear.

## Common Commands

### Search and Jobs

```bash
splunkctl search 'index=main error' --output json
splunkctl search 'index=main sourcetype=access_combined status=500' --maxout 50 --output json
splunkctl search export 'index=main | head 100' --output json

splunkctl search 'index=main earliest=-7d' --detach
# returns {"job_id":"<sid>","ttl":3600}
splunkctl jobs status <sid> --output json
splunkctl jobs show <sid> --output json
splunkctl jobs list --output json
```

### Indexes

```bash
splunkctl index list --output json
splunkctl index list --name myindex --output json
splunkctl index add --name newindex
splunkctl index disable myindex
splunkctl index enable myindex
```

### Users and Roles

```bash
splunkctl user list --output json
splunkctl user add jdoe --role admin --password '<new-user-password>'
splunkctl user edit jdoe --roles power
splunkctl user password jdoe --password '<new-password>'
splunkctl role list --output json
```

### Apps, Health, and Cluster

```bash
splunkctl app display --output json
splunkctl server health --output json
splunkctl health list --output json
splunkctl cluster-peers list --output json
splunkctl shcluster-status show --output json
```

### Knowledge Objects and Alerts

```bash
splunkctl saved-search list --output json
splunkctl saved-search add "My Alert" --search 'index=main error | stats count' --param cron_schedule="0 * * * *" --param is_scheduled=1
splunkctl fired-alert list --output json
splunkctl lookup list --output json
splunkctl macro list --output json
splunkctl eventtype list --output json
splunkctl field-extraction list --output json
splunkctl data-model list --output json
```

### HEC

```bash
splunkctl hec list --output json
splunkctl hec add mytoken --index main --sourcetype json
splunkctl hec send --hec-token '<hec-token>' --index main --event '{"message":"test"}'
```

## Parameters and Output

- Use `--output json` for structured resource commands.
- Use `--describe` before `--param` when adding or editing resources:

```bash
splunkctl index add --describe
splunkctl user add --describe
splunkctl saved-search add --describe
```

- Errors are JSON on stderr: `{"error":"...","code":"NOT_FOUND"}`.
- Exit codes: `0` success, `1` usage/general fallback, `3` not found, `4` auth error, `5` connection error.
- Resource-first command shape is required: use `splunkctl index list`, not `splunkctl list index`.
