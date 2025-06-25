# AGENTS.md

A single, authoritative operating manual for **humans and AI agents (e.g. CODEX in ChatGPT)** who contribute to this repository.

---
## 1. Mission
Fix, test and merge every documented bug (`docs/bugs/**/*.md` and `docs/bugs-phase-02/**/*.md`).  Each bug is resolved in its own branch and verified by automated CI.

---
## 2. Repository Rules
1. **Never** commit directly to `main` or `dev-*` branches.
2. Each bug ⇒ dedicated feature branch: `bug/<bug-id>-<kebab-title>`  
   Example: `bug/03-race-condition-websocket-handler`.
3. Keep documentation in sync with code (update bug index, status fields, counters).
4. Create a resolved-bug report under `docs/bug-fix/` once the fix is merged.  
   File name format: `<bug-id>-<kebab-title>-resolved.md` (e.g. `01-bug-01-missing-proto-generation-resolved.md`).
5. **Trigger phrase**: asking `RESOLVE THE BUG <bug-file>` kicks off the automated fix workflow.
5. Squash-merge every PR via GitHub UI.

---
## 3. Development Workflow (TL;DR)
```bash
# 0. Pre-requisites (first time only)
make deps proto            # download modules & generate code
make ci                    # run local mirror of CI pipeline (works offline)

# 1. Pick a bug & create branch
BUG=03-bug-03-race-condition-websocket-handler
BRANCH="bug/${BUG%%-*}-${BUG#*-}"
git checkout -b "$BRANCH"

# 2. Reproduce failure with a new test (should RED)
#    ‑ add *_test.go reproducer per bug report instructions
make test -race | cat

# 3. Implement the fix & update docs
#    ‑ follow Recommended Fix in markdown file
#    ‑ update Status in bug file → Fixed
#    ‑ run `make proto` if proto files changed
#    ‑ create resolved report in docs/bug-fix/<bug-id>-<title>-resolved.md

# 4. Verify locally
make ci                    # build, vet, lint, tests, race, proto-diff (works offline)

# 5. Commit & push
git add .
git commit -s -m "fix(websocket): race in broadcast map (03-bug-03)"
git push --set-upstream origin "$BRANCH"

# 6. Open PR (template auto-loads)
#    PR title:  "fix: 03-bug-03 – race in websocket broadcast"
```

Detailed per-bug steps are documented in section 8.

---
## 4. Pull-Request Checklist (enforced)
- [ ] Links to bug markdown (e.g. `closes docs/bugs/03-bug-03-race-condition-websocket-handler.md`).
- [ ] `make ci` passes (build + vet + staticcheck + tests ‑race + proto diff clean) — works offline.
- [ ] New or updated tests cover the fix (failing on `dev` / passing on branch).
- [ ] Documentation updated (bug Status, overview counters, CHANGELOG if needed).
- [ ] Resolved report added under `docs/bug-fix/` with verification steps.
- [ ] At least one reviewer approval.

---
## 5. Coding Standards
* Go **1.22**.  Run `goimports` on save.
* **Concurrency**: no map writes under `RLock`; always run `go test -race`.
* Return `%w` wrapped errors; log using `zap` (`debug|info|warn|error`).
* No TODOs committed to `main`.

---
## 6. Testing Standards
* All fixes require a failing test first (Red-Green-Refactor).
* Race detector and coverage (≥70 %) run in CI.
* Integration tests live under `tests/`.

---
## 7. Branch & Commit Conventions
| Item   | Format | Example |
|--------|--------|---------|
| **Branch** | `bug/<id>-<kebab-title>` | `bug/05-configuration-port-mismatch` |
| **Commit** | `fix(<area>): <short> (#bug-id)` | `fix(websocket): unsubscribe logic error (#12)` |
| **PR title** | `fix: <bug-id> <short>` | `fix: 09-bug-09 race in broadcast` |

---
## 8. Per-Bug Operating Procedure (expanded)
1. **Branch** – `git checkout -b bug/<id>-<title>`.
2. **Test First** – translate reproduction steps into failing test.
3. **Verify RED** – `make test -race`; ensure failure.
4. **Fix** – implement code & docs changes.
5. **Local CI** – `make ci` (works offline).
6. **Docs** – mark bug as `Fixed`, update overview counts.
7. **Push & PR** – PR auto-fills template; complete checklist.
8. **Review & Merge** – squash merge once green.
9. **Branch Cleanup** – delete remote branch.

---
## 9. Tooling Commands (quick reference)
| Purpose      | Command                     |
|--------------|----------------------------|
| Install deps | `make deps`                |
| Generate PB  | `make proto`               |
| Build app    | `make build`               |
| Full CI      | `make ci` (offline ready)  |
| Hot reload   | `make dev`                 |
| Coverage     | `make test-coverage`       |

---
## 10. Automation Matrix
| Stage  | Trigger | Agent / Tool               |
|--------|---------|----------------------------|
| CI     | push/PR | GitHub Actions (`ci.yml`)  |
| Review | PR      | Human / ChatGPT-CODEX      |
| Merge  | PR green & approval | GitHub UI |

---
## 11. Fallback / Escalation
If CI fails with a non-trivial race or flaky test, label the PR `needs-investigation` in a comment. 

---
## 12. Bug-Fix Documentation (`docs/bug-fix/`)

Every resolved bug **must** include a standalone markdown report located at `docs/bug-fix/<bug-id>-<kebab-title>-resolved.md`.

**Required sections:**

| Section | Purpose |
|---------|---------|
| Header  | Copy original bug header (ID, Severity, Date) + add `Status: Resolved` |
| Summary | Short description of root cause & fix approach |
| Code Changes | Bullet list of affected files / functions |
| Tests Added | Paths to new/updated `*_test.go` files |
| Verification | CLI commands to reproduce failure on `dev` & success on branch |
| Checklist | Tick-box PR checklist (mirrors Section 4) |

Include any diagrams, logs, or links that help reviewers understand the fix.  Place binary assets (e.g. PNG) in `docs/bug-fix/assets/`.

---
### Deep-Wiki & API Docs locations

* Internal architecture docs live under `docs/deepwiki-docs/`
* External API reference snapshots live under `docs/delta-exchange-api-docs/`

Refer to these paths when cross-linking from resolved reports. 

---
## 13. Telemetry Logging Implementation Workflow

The observability roadmap lives in `docs/telemetry-planning.md`.  Each **phase** (P-1 … P-7) is delivered with the same discipline used for bug fixes, but through a dedicated trigger phrase so that CODEX can automate scaffolding and quality-gates.

### 13.1 Trigger Phrase
> **IMPLEMENT THE PHASE-<n> OF LOGGING in docs/telemetry-planning.md**

Using the phrase above (replace `<n>` with 1-7) in a PR description or chat message immediately spins up the following automated workflow:
1. Create a feature branch `telemetry/phase-<n>`.
2. Apply code changes listed for that phase in `docs/telemetry-planning.md`.
3. Run `make ci` (works offline) and telemetry smoke-tests.
4. Open a draft PR ready for review.

### 13.2 Per-Phase Checklist (CI Enforced)
- [ ] Code matches exact insertion points & items in `docs/telemetry-planning.md`.
- [ ] All existing tests + new telemetry tests pass (`make ci` — works offline).
- [ ] `/metrics` endpoint returns **HTTP 200** and exposes new metrics family.
- [ ] Jaeger (or Tempo) shows at least one span emitted by this phase (smoke test).
- [ ] Feature flag `TELEMETRY_PHASE_<n>_ENABLED` defaults **true** in dev, **false** in prod charts until reviewed.
- [ ] Documentation — toggle corresponding checkbox in Section 6 of `telemetry-planning.md`.

### 13.3 Branch & Commit Convention
| Item   | Format                                    | Example                               |
|--------|-------------------------------------------|---------------------------------------|
| Branch | `telemetry/phase-<n>-<short-title>`       | `telemetry/phase-3-broadcast-hub`     |
| Commit | `telemetry(<area>): phase-<n> <summary>`  | `telemetry(websocket): phase-2 ingest`|
| PR     | `telemetry: phase-<n> – <short-title>`    | `telemetry: phase-1 core bootstrap`   |

Follow the existing review, merge and escalation rules (Sections 4, 10 & 11) for all telemetry work. 

---
## 14. Offline Builds & CODEX Compatibility

Since CODEX (ChatGPT) environments don't have internet access, the repository includes a `vendor/` directory with all Go dependencies for offline builds.

### 14.1 How It Works
- **With internet** (local development): `make ci` downloads dependencies as needed and runs full CI pipeline
- **Without internet** (CODEX): `make ci` automatically uses vendored dependencies with `-mod=vendor` and runs the same full CI pipeline

### 14.2 Maintaining Vendor Directory
When dependencies change, update the vendor directory:
```bash
make vendor
git add vendor/ go.mod go.sum
git commit -m "vendor: update dependencies for CODEX compatibility"
```

### 14.3 Makefile Auto-Detection & Protobuf Compatibility
The `Makefile` automatically detects if `vendor/` exists and switches to offline mode:
- ✅ **Offline mode**: Uses `-mod=vendor` flags when `vendor/` directory exists
- 📥 **Online mode**: Downloads dependencies when `vendor/` directory is missing

**Protobuf Version Compatibility:**
- `proto-check` target ignores protoc version differences in generated file comments
- Local dev (protoc v5.29.3) vs CODEX (protoc v5.27.1) differences are automatically ignored
- Only actual code changes in protobuf files will cause CI failures

This ensures both local development and CODEX environments work seamlessly.

---