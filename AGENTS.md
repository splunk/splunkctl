# splunkctl — Agent Working Guide

This file is the normative entry point for any AI agent working on this repository. Read it before changing code, generated output, tests, documentation, or build tooling. The topic documents below explain detailed workflows; when contributor guidance conflicts, follow this file and fix the contradiction in the same change when it is in scope.

## What this is

`splunkctl` is a standalone Go CLI and module (`github.com/splunk/splunkctl`) for managing Splunk through its REST API.

It is **not** a replacement for the `splunk` binary. It covers the REST-backed command surface defined by the canonical `splunkrc_cmds.xml`; that source file may live outside a standalone checkout. It does not provide `start`, `stop`, `diag`, or other local-filesystem commands.

## Requirement language and MR gate

- **MUST** and **MUST NOT** are hard gates. An agent MUST NOT create an MR until every applicable requirement is satisfied.
- **SHOULD** and **SHOULD NOT** state the required default. A deviation MUST have a concrete, repository-specific justification in the MR; preference or time pressure is not enough.
- Unqualified imperatives such as “do not” and “never” have **MUST** or **MUST NOT** force. Only **SHOULD**, **SHOULD NOT**, and **MAY** express permitted discretion.
- A check that does not apply is not an exception. Record why it does not apply when that would not be obvious to a reviewer.

Only one kind of exception may be considered: an applicable required check is unavailable because of an external or environmental prerequisite outside the change. Before creating an MR, the agent MUST:

1. record the exact check that was not run;
2. identify the unavailable prerequisite and why it could not be supplied;
3. provide the substitute evidence that was obtained;
4. state the residual risk and a concrete follow-up; and
5. obtain the user's explicit approval for that specific exception.

A failed correctness, security, data-loss, compatibility, or required-validation check is not an exception and MUST be fixed. An agent MUST NOT state or imply that an unrun check passed. No other `MUST` may be waived under this policy.

## Applicability and pre-existing debt

Applicability is scope determination, not an exception. A requirement or finding is applicable when it concerns any of the following:

- prospective new or changed code, tests, generated output, documentation, configuration, or build artifacts;
- a contract or reachable behavior whose outcome can change because of the prospective change; or
- a dependency, assumption, or validation/safety evidence on which the change relies.

Unrelated pre-existing debt alone does not block an MR. It MUST NOT be concealed, copied into new work, or worsened, and it MUST be recorded when it is relevant to understanding or reviewing the change. Pre-existing debt becomes in-scope and blocking when the change relies on it, newly exposes users or callers to it, or worsens it; makes its later repair harder; or prevents the required safety or validation evidence from being obtained.

## Docs index

| File | What it covers |
|---|---|
| `docs/architecture.md` | Package layout, data flow, key invariants |
| `docs/codegen.md` | How the XML → Go generator works and how to run it |
| `docs/adding-commands.md` | How to add a generated command or override |
| `docs/testing.md` | Test structure, commands, and test-writing conventions |
| `docs/decisions.md` | Durable design choices and their rationale |
| `.gitlab/merge_request_templates/Default.md` | Concise MR evidence prompts and AI-assistance disclosure |

## Stable CLI invariants

- **Command shape:** commands MUST be resource-first: `splunkctl <object> <verb>`. They MUST NOT use `verb object`.
- **Authentication:** host and token are the only authentication settings. Each resolves in this order: command-line flag (`--host`/`--token`) → environment (`SPLUNKCTL_HOST`/`SPLUNKCTL_TOKEN`) → `~/.splunkctl/config.yaml`. Username/password authentication MUST NOT be added.
- **Override models:** the preferred model is a full-object replacement: a same-name `cmd/override/<object>.go` makes codegen skip that XML object. Codegen does not delete a previously emitted object file, so a full replacement MUST remove that stale artifact intentionally and update registration. The repository also has explicit legacy partial overlays for `app`, `index`, `user`, and `scheduler` in `cmd/register_overrides.go`; these attach or replace verbs on checked-in generated parents and are not automatic codegen behavior.
- **Exit codes:** preserve the current contract: 0=success; 1=general or usage fallback (`USAGE_ERROR` for the fallback); 3=NOT_FOUND; 4=AUTH_ERROR; and 5=CONNECTION_ERROR. Exit 2 is reserved and currently unused. Changing usage failures to exit 2 is an intentional compatibility change and requires scoped implementation, tests, help/schema review, documentation, and MR disclosure.
- **Output contracts:** structured resource/REST results MUST use `output.Print(cmd.Context(), data, format)`. Raw or fixed-format responses, interactive prompts, and progress displays MAY use Cobra writers or an established terminal-specific path when that is their explicit contract. New direct writes to `os.Stdout`/`os.Stderr` or calls to `fmt.Print*` MUST NOT be added unless an established special path requires the process stream and tests cover that behavior. The established centralized root error handler MUST emit stderr JSON containing only `error` and `code`; it MAY continue to use its process stream.

## Package ownership and dependency direction

Keep each responsibility at its existing boundary:

- `cmd/` owns root-command lifecycle, persistent flag/config wiring, registration, shared Cobra wiring, and top-level commands.
- `cmd/generated/` contains emitted object adapters and hand-maintained support/tests. Only object files carrying the standard `// Code generated by splunkctl codegen. DO NOT EDIT.` header are generated artifacts; `shared.go`, `_test.go` files, and `.gitkeep` are hand-maintained.
- `cmd/override/` owns hand-written Cobra adapters and command-specific orchestration that cannot be expressed by the generator, including full-object replacements and the explicitly registered legacy partial overlays.
- `internal/client/` owns Splunk REST transport, authentication headers, request/response handling, and transport/API error representation.
- `internal/output/` owns reusable structured-result rendering and format selection. It MUST NOT acquire transport or command-orchestration responsibilities.
- `internal/config/` owns configuration resolution.
- `internal/cmdctx/` is the shared context/dependency bridge used to avoid package cycles.
- `codegen/` owns XML parsing, the source transformation, generation policy, and templates.

Dependencies MUST continue to point from command adapters toward focused internal capabilities. Lower-level internal packages MUST NOT import `cmd`, `cmd/generated`, or `cmd/override`. The `cmd` package imports `cmd/generated` for registration, so `cmd/generated` MUST NOT import back into `cmd`; shared client access belongs in `internal/cmdctx`. This invariant MUST NOT be bypassed with copied context keys, package globals, or a new dependency back-edge.

See `docs/architecture.md` for the full data flow.

## Engineering standards

### Cobra handlers and cohesive behavior

A `RunE` handler MUST remain a thin adapter. It MUST:

1. parse and validate arguments and flags;
2. obtain context-injected dependencies;
3. invoke one cohesive behavior while propagating `cmd.Context()`; and
4. render structured results with `output.Print` or use the command's established raw, fixed-format, interactive, or progress-output contract.

Simple command-specific behavior MAY remain in the adapter. Extraction MUST be limited to behavior that is genuinely cohesive, multi-step, or independently testable. When extraction is justified, use a well-named function in the owning package or a focused `internal/<capability>` package with a clear present responsibility.

New packages MUST have a precise owner and reason to exist now. Generic `util`, `common`, `misc`, `helpers`, `types`, or `interfaces` dumping grounds MUST NOT be created. If no real package boundary exists, cohesive files SHOULD be split within the existing owning package.

There are no arbitrary function- or file-length thresholds. Organization decisions MUST be based on responsibility, dependency direction, testability, and change coupling.

### Reuse, interfaces, dependencies, and state

- Reuse existing code when it implements the same semantics, invariants, error behavior, and contract—not merely because two fragments look alike.
- Speculative abstractions and mechanical DRY refactors MUST NOT be introduced. If likely callers or evolution differ, duplication MAY be clearer than false coupling.
- Interfaces SHOULD be small, owned by the consumer, and introduced only for a demonstrated substitution or test boundary. Constructors SHOULD normally return concrete types.
- The Go standard library SHOULD be preferred. A new dependency MUST have a documented need and a review of its maintenance, security, binary-size, and licensing impact as applicable.
- Dependencies MUST be explicit in parameters, constructors, or the established context-injection path. They MUST NOT be hidden behind new mutable runtime globals.
- New mutable runtime globals MUST NOT be introduced unless an unavoidable invariant and lifecycle are documented and tested.

### Context, errors, and compatibility

- Propagate `cmd.Context()` through all request-path and blocking work. Do not replace it downstream with `context.Background()`, store request contexts in structs, or drop cancellation and deadlines.
- Return errors to Cobra instead of printing-and-continuing. Add useful operation and resource context and preserve unwrap behavior, normally with `%w`.
- Error text and logs MUST NOT expose tokens, authorization headers, credentials, secret configuration, or sensitive response bodies.
- Preserve existing command names, resource-first shape, flags, environment variables, output schemas/formats, error semantics, and exit codes unless the task explicitly changes the contract. Intentional compatibility changes MUST be called out in help/schema, tests, documentation, and the MR.

## Generated code

Within `cmd/generated/`, only emitted object files carrying the standard `// Code generated by splunkctl codegen. DO NOT EDIT.` header are generated artifacts. Those emitted files MUST NOT be hand-edited. `shared.go`, `_test.go` files, and `.gitkeep` are hand-maintained and MUST be reviewed and edited normally when their contracts change.

Changes to an emitted object file MUST be made at its source:

- update the canonical XML when it is available and the REST command definition is changing;
- update the parser, generator, or template when generation behavior is changing; or
- use a full-object override when the object needs hand-written behavior.

A full-object override is the required default for new override work. A same-name override causes codegen to skip the XML object, but the generator never deletes a previously emitted file. The stale emitted object file MUST be removed intentionally and `cmd/register_generated.go` and `cmd/register_overrides.go` MUST be updated consistently.

The existing `app`, `index`, `user`, and `scheduler` partial overlays are explicit legacy wiring in `cmd/register_overrides.go`: they attach or replace override verbs on checked-in generated parents even though their same-name override files make future generator runs skip those objects. The `app` overlay replaces generated `install` and adds `filename=true` for file-path installs. They MUST NOT be treated as automatic generation behavior. A new partial overlay MUST NOT be introduced without an explicit design, registration behavior tests, a plan for generator and emitted-artifact ownership, and documentation. A change to an existing overlay MUST audit the generated parent, override verbs, registration, and future regeneration behavior together.

The emitted files MUST be regenerated and committed together with the source change. Follow the authoritative generator validation under [Required final-revision validation](#required-final-revision-validation); see `docs/codegen.md` for the explicit-input workflow and prerequisites.

The canonical XML is an external prerequisite in a standalone checkout. If the required revision is unavailable, the agent MUST stop, disclose the missing prerequisite, and follow the exception process before any MR. The agent MUST NOT fabricate XML, claim regeneration occurred, or patch emitted files as a substitute.

## Tests

Tests MUST exercise the layer whose contract changed.

- Transport, authentication, response parsing, and API-error changes belong in `internal/client` tests.
- A client-only test is insufficient for a Cobra command change. Test realistic arguments and flags, command registration or the responsible handler, context-injected dependencies, and captured output as applicable.
- New commands SHOULD expose a package-local constructor that returns a fresh Cobra command or a cohesive handler that a fresh test command can invoke. Tests SHOULD avoid mutating package-global command trees; when existing globals must be used, isolate and restore their state and do not mutate them from parallel tests.
- Use `httptest.Server` for HTTP behavior. Unit and command tests MUST NOT require a real Splunk deployment, real credentials, or external network connectivity; loopback traffic to an in-process `httptest.Server` is allowed.
- Cover relevant success, API/transport failure, invalid input, empty and boundary cases, and cancellation/deadline behavior. “Relevant” is determined by the changed contract, not by a mechanical quota.
- Test helpers MUST call `t.Helper()`. Keep fixtures focused, failures diagnostic, and tests deterministic; tests are maintainable production assets, not disposable scaffolding.

Every command change MUST include focused tests. Follow the authoritative command smoke checks under [Required final-revision validation](#required-final-revision-validation); see `docs/testing.md` for repository test helpers and workflows.

Integration tests under `tests/integration/` run live Splunk checks separately from hermetic unit and command tests. Run `make test-integration` after `make build`, or use `make test-integration-local` for local provisioning. Regenerate golden files only intentionally against a clean state; local bootstrap credentials are used only to create temporary bearer tokens and MUST NOT be committed.

## Local reproducibility

Local development is part of the contract:

- Documented runnable workflows MUST be copy-pasteable from the repository root. An environment-specific substitution MAY be required only when it is clearly identified and defined adjacent to the command that uses it.
- A command template containing a substitution MUST NOT be reported as literal executed evidence. Evidence MUST identify the resolved input and actual command form, with secrets redacted.
- Prerequisites MUST be explicit. Do not rely silently on a sibling checkout, inherited credentials, uncommitted machine-local generated state, a globally installed tool, or other machine-specific state.
- Build, unit-test, root help, version, and schema workflows MUST remain runnable without a live Splunk service or credentials. Dependency download is a separately documented prerequisite when the module cache is empty.
- Document every new tool, service, environment variable, fixture-generation step, and platform assumption in the appropriate workflow document.

The authoritative local validation commands and change-specific checks are centralized under [Required final-revision validation](#required-final-revision-validation).

If a normal repository-root workflow is broken, the change is not ready for an MR. The implementation or documentation MUST be updated; future agents MUST NOT be left to rediscover a hidden setup step.

## Documentation is part of the change

For every change, identify affected documentation surfaces and update them in the same MR:

| Change | Required documentation surface |
|---|---|
| User workflows, installation, prerequisites, or local setup | `README.md` |
| Package ownership, dependencies, or data flow | `docs/architecture.md` |
| Generator inputs, behavior, or regeneration workflow | `docs/codegen.md` |
| Command-extension mechanics or override rules | `docs/adding-commands.md` |
| Test commands, fixtures, layers, or conventions | `docs/testing.md` |
| Durable architectural trade-offs or compatibility decisions | `docs/decisions.md` |
| Pre-MR gate or required MR evidence | `.gitlab/merge_request_templates/Default.md` |
| Command names, flags, arguments, defaults, output, or exit behavior | Cobra help and schema metadata |
| New or materially changed exported declarations | Go doc comments following Go conventions |
| New packages | A package comment explaining the package responsibility |
| Non-obvious invariants or constraints in code | A focused inline comment explaining **why**, not the obvious **what** |

Search all relevant surfaces, including `README.md`, `DEMO.md`, `docs/`, Cobra help/schema, examples, and embedded skills such as `cmd/override/skill_*.md`. Inline comments MUST NOT replace user or workflow documentation. Remove stale instructions instead of layering a new contradictory note on top.

If documentation is unaffected, the MR MUST say why after this impact review.

## Mandatory pre-MR self-review

After implementation is complete and before requesting or creating an MR, the authoring agent MUST stop implementation work and perform a separate critical review of the final prospective MR.

### Establish the exact review surface

1. Identify the actual target remote and branch; `origin` and `main` MUST NOT be assumed.
2. Refresh that target from its remote, record the exact target commit, and compute the merge base from that refreshed commit. If the target cannot be refreshed because of an external or environmental prerequisite, use the exception process before creating an MR; a stale local ref is not fresh evidence.
3. Inspect the prospective MR's stat, whitespace check, and full diff from that merge base.
4. Inspect staged changes, unstaged changes, and every untracked file separately. Untracked files do not appear in a normal diff.
5. Confirm the final scope contains only intentional files and that generated artifacts are traceable to their sources.

Use the equivalent of:

```bash
(
    set -eu
    target_remote="REPLACE_WITH_TARGET_REMOTE"
    target_branch="REPLACE_WITH_TARGET_BRANCH"
    target_ref="refs/remotes/${target_remote}/${target_branch}"
    git fetch --no-tags "$target_remote" \
        "+refs/heads/${target_branch}:${target_ref}"
    target_commit="$(git rev-parse --verify "${target_ref}^{commit}")"
    merge_base="$(git merge-base HEAD "$target_commit")"
    printf 'target_ref=%s target_commit=%s merge_base=%s\n' \
        "$target_ref" "$target_commit" "$merge_base"
    git status --short
    git diff --stat "$merge_base"
    git diff --check "$merge_base"
    git diff "$merge_base"
    git diff --cached
    git diff
    git ls-files --others --exclude-standard
)
```

Set `target_remote` and `target_branch` to the actual MR target before running the block; the `REPLACE_WITH_...` values are a safe template, not literal evidence. The fully qualified refspec fetches that selected branch into its exact remote-tracking ref even when the clone has a narrowed fetch configuration. The subshell is fail-fast: a failed fetch, ref verification, merge-base calculation, or diff check stops the block. Record the resolved target ref, exact refreshed commit, merge base, fetch command/result, and diff commands. The merge-base diff includes committed branch work and tracked working-tree changes; the separate status and untracked-file review closes the remaining gaps. Open every listed untracked file and review its full contents.

### Read and challenge the change

The authoring agent MUST personally read every human-written changed line and enough surrounding callers, callees, tests, and documentation to judge the change. The review MUST verify:

- every user requirement maps to implementation and evidence;
- applicability covers the prospective artifacts, directly affected or reachable contracts, relied-on dependencies/evidence, and any relevant pre-existing debt;
- the MR has one cohesive purpose and no unrelated refactor or drive-by change;
- correctness across success, failure, boundary, cancellation, and compatibility behavior;
- errors preserve useful context without leaking secrets;
- package ownership, dependency direction, modularity, semantic reuse, and future maintainability;
- security, data-loss, reliability, and performance implications;
- generated-file provenance, a stable non-sensitive embedded input label, reproducible content, and a complete emitted-artifact inventory;
- tests cover the changed boundary and would fail for the important regressions;
- documentation, examples, help/schema, and local-run instructions agree with the code; and
- no unintended debug output, temporary files, unexplained generated files, tracked build artifacts, credentials, dead code, stale comments, or conflict markers remain.

Automated tools, an independent reviewer, and delegated agents can add evidence, but they MUST NOT replace the authoring agent's personal diff reading and judgment.

## Required final-revision validation

Run these exact baseline commands from the repository root after the last implementation edit:

```bash
git diff --check
make fmt-check
go vet ./...
make test
make build
go run . --help
go run . version --output json
go run . schema --groups
```

Every command MUST succeed on the same revision proposed for review, subject only to the explicit unavailable-prerequisite exception process above. Record the exact commands and results; an unrun command MUST NOT be summarized as covered by another command.

Run these additional checks when applicable:

- **Concurrency, cancellation, or race-sensitive change:** `go test -race ./...` or the narrowest command that exercises all affected packages.
- **Dependency change:** `go mod tidy -diff` and `govulncheck ./...`; review `go.mod` and `go.sum` for intended-only changes. If the vulnerability tool is unavailable, use the exception process rather than claiming a scan passed.
- **Performance claim or hot-path change:** run the relevant reproducible benchmark before and after the change and report the command and comparison.
- **Code generation, code-generation template, XML, override-ownership, or generated-output change:** record the immutable compatible splcore source revision and SHA-256 digest of `splunkrc_cmds.xml`. Use a stable, non-sensitive XML argument because codegen embeds that argument verbatim in every emitted header. Run the documented explicit-input generation command, audit the complete emitted-artifact inventory, intentionally remove stale emitted files and update registration when objects disappear or become full overrides, snapshot the complete `cmd/generated/` tree, then run the identical command again and compare the full output tree with that snapshot. A clean second run proves output-tree idempotence only; it does not prove that the artifact set is correct.
- **Command change:** define `object="REPLACE_WITH_OBJECT"` and `verb="REPLACE_WITH_VERB"`, then run focused tests plus `go run . "$object" "$verb" --help`; also smoke any offline behavior whose contract changed. Record the resolved command path, not the template values.

After any review fix, rerun every affected check and reread the resulting final diff. Evidence from a superseded revision is not final evidence.

## Required MR evidence

Start from `.gitlab/merge_request_templates/Default.md` for every MR and explicitly apply it in CLI or API creation flows. Keep the description concise by grouping related evidence or linking to a durable design or validation record, but do not omit applicable evidence from the list below. Complete every visible field, retain and complete the AI-assistance disclosure, and give a concrete reason for each non-obvious `N/A`. The template records evidence; it does not replace the mandatory personal self-review above.

The MR description MUST include:

- what changed and why;
- a mapping from requirements to implementation and tests;
- design choices, rejected alternatives, and maintainability trade-offs where material;
- compatibility and user-visible impact;
- target remote and branch, successful refresh command/result, exact target commit, and merge base;
- exact validation commands and results from the final revision;
- documentation updated, or a concrete explanation of why none was affected;
- generated-file provenance—including compatible source revision, XML SHA-256 digest, stable input label, emitted-artifact inventory, stale-file decisions, and complete output-tree second-run comparison evidence—when applicable;
- risk assessment covering security, data loss, reliability, and performance as applicable;
- relevant pre-existing debt, its applicability determination, and why it is blocking or non-blocking; and
- known limitations, follow-up work, and every explicitly approved exception with its approval context.

An agent MUST NOT create an MR while any applicable, in-scope correctness, security, data-loss, compatibility, maintainability, test, documentation, local-run, or required-validation finding remains unresolved. The database may forgive optimism; the review gate will not.

## Research basis

These rules adapt the repository-specific design in `docs/plans/2026-08-10-pre-mr-engineering-standards-design.md` and established review and Go guidance: [Google engineering practices](https://google.github.io/eng-practices/review/), [GitLab code review guidelines](https://docs.gitlab.com/development/code_review/), [Go module organization](https://go.dev/doc/modules/layout), [Go package names](https://go.dev/blog/package-names), and [Go doc comments](https://go.dev/doc/comment).
