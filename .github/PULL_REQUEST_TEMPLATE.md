# Pull Request

> **IMPORTANT**: This PR must address exactly **one** documented bug.

## Linked Bug
Closes <!-- docs/bugs/xx-bug-xx-*.md or #IssueNumber -->

## Checklist
- [ ] Branch named `bug/<id>-<kebab-title>`
- [ ] `make ci` passes locally and in GitHub Actions
- [ ] Failing test reproduced on `dev` and now passes
- [ ] New / updated tests cover the fix (lines / functions)
- [ ] `staticcheck`, `go vet`, `go test -race` all green
- [ ] Generated protobuf code up-to-date (`make proto && git diff --exit-code`)
- [ ] Documentation updated:
  - [ ] Bug markdown **Status:** Fixed
  - [ ] Counters in `docs/bugs/00-overview_of_bugs.md`
- [ ] PR title follows convention: `fix: <bug-id> <short>`
- [ ] Squash merge selected

## Description of Fix
<!-- High-level summary: what was wrong & how it was fixed. Reference code sections if helpful. -->

## Testing Strategy
<!-- Describe how you verified the fix beyond automated tests (e.g. manual steps, load tests). -->

## Screenshots / Logs (optional) 