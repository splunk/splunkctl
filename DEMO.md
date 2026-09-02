# splunkctl Agentic Demo

## Before you start

**1. Build and install the CLI**

From the `splunkctl` repository root:

```bash
make build
./bin/splunkctl --help

# Make the repository build available as `splunkctl` in this shell.
export PATH="$PWD/bin:$PATH"
```

Alternatively, copy `bin/splunkctl` to a directory already on your `PATH`.

**2. Configure it**
```bash
splunkctl config set host https://localhost:8089
splunkctl config set token <your-token>
```
Get a token: Splunk Web → Settings → Tokens → New Token.

**3. Install the agent skill**
```bash
splunkctl skill install --interactive
```
Choose `claude` and `local`. This writes `.claude/skills/splunkctl/SKILL.md` — Claude picks it up automatically.

**4. Open a Claude Code session in this directory**
```bash
claude
```

---

## The Demo

Paste these prompts one at a time into your Claude Code session.
Each one exercises a different part of the CLI. Watch Claude discover and run the right command.

---

### Health & Status

> What's the current health of my Splunk instance? Any warnings or issues?

---

> Show me the current server info — version, OS, host, and license type

---

> What's the pipeline queue fill rate? Are any queues backing up?

---

### Search

> Show me the top 10 sourcetypes by event count in the last hour

---

> Run a background search for errors in `index=_internal` grouped by sourcetype over the last hour — I'll check the results in a moment

---

> Show me the results for that job

---

### Indexes

> List all indexes with their current sizes and retention settings

---

> Create an index called `audit-logs` with a 90-day retention period

---

> Change the retention on `audit-logs` to 180 days

---

> Delete the `audit-logs` index

---

### Users & Access Control

> Who has admin access to this Splunk instance?

---

> Create a new user called `mchen` with the `power` role, email `mchen@example.com`, password `TempPass123!`

---

> Reset mchen's password to `NewSecure456!`

---

> Remove the user mchen

---

### HTTP Event Collector

> List all HEC tokens and show which ones are enabled

---

> Create a new HEC token called `ingest-hec` on index `main` with sourcetype `app_events`

---

> Send a test event through the `ingest-hec` token: `{"message": "splunkctl demo", "level": "info"}`

---

> Delete the `ingest-hec` HEC token

---

### Saved Searches & Alerts

> List all saved searches — which ones are scheduled?

---

> Create a saved search called `Error Spike Monitor` that runs every 5 minutes: `index=_internal log_level=ERROR | stats count | where count > 100`

---

> Dispatch `Error Spike Monitor` right now and show me the results

---

> Are there any fired alerts in the system?

---

### Dashboards

> List all dashboards — which ones are shared with everyone?

---

> Create a dashboard called `soc-overview` that shows a timechart of errors from `index=_internal log_level=ERROR`

---

> Update the `soc-overview` dashboard to also show a table of the top 5 error messages

---

> Make the `soc-overview` dashboard readable by everyone but only writable by admins

---

> Delete the `soc-overview` dashboard

---

### KV Store

> Create a KV store collection called `event_store` in the `search` app

---

> Add a record to `event_store`: `{"session_id": "abc123", "user": "mchen", "ts": "2025-01-01"}`

---

> Query all records in `event_store`

---

> Delete the `event_store` collection

---

### Configuration & Apps

> What's in the `[default]` stanza of `server.conf`?

---

> List all installed apps — which ones are disabled?

---

### License

> How much of my daily license quota am I using?

---

> List the license pools and how much of each is consumed

---

### Escape Hatch

> Call the raw Splunk REST API to get `/services/server/settings`

---
