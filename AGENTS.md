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
4. Squash-merge every PR via GitHub UI.

---
## 3. Development Workflow (TL;DR)
```bash
# 0. Pre-requisites (first time only)
make deps proto            # download modules & generate code
make ci                    # run local mirror of CI pipeline

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

# 4. Verify locally
make ci                    # build, vet, lint, tests, race, proto-diff

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
- [ ] `make ci` passes (build + vet + staticcheck + tests ‑race + proto diff clean).
- [ ] New or updated tests cover the fix (failing on `dev` / passing on branch).
- [ ] Documentation updated (bug Status, overview counters, CHANGELOG if needed).
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
5. **Local CI** – `make ci`.
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
| Full CI      | `make ci`                  |
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