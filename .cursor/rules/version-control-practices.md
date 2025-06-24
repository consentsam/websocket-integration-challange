# Version Control Practices

These rules ensure consistent, traceable, and professional version control practices for the WebSocket service project.

## Commit Message Standards

### Conventional Commit Format
```
type(scope): brief description

Optional longer description explaining what changed and why.

- Bullet points for multiple changes
- Reference issues with Fixes #123 or Closes #123
- Include breaking change notes if applicable

BREAKING CHANGE: Description of breaking change if applicable
```

### Commit Types
- **feat**: New features or functionality
- **fix**: Bug fixes
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code restructuring without changing functionality
- **test**: Adding or modifying tests
- **chore**: Maintenance tasks, dependency updates, etc.
- **perf**: Performance improvements
- **ci**: CI/CD configuration changes

### Examples
```
feat(websocket): add client subscription management

Implement subscription tracking for WebSocket clients including:
- Channel-based subscription filtering
- Product ID filtering for ticker data
- Automatic cleanup on client disconnection

Fixes #42
```

```
fix(delta-client): handle connection timeout gracefully

Add proper error handling and retry logic when Delta Exchange
connection times out. Previously would panic, now logs error
and attempts reconnection with exponential backoff.

Fixes #38
```

```
docs(api): update WebSocket message format documentation

- Add examples for all message types
- Clarify error response format
- Update authentication requirements

Closes #45
```

## Branch Naming Conventions

### Branch Types
- **feature/**: New features (`feature/websocket-auth`)
- **bugfix/**: Bug fixes (`bugfix/connection-leak`)
- **hotfix/**: Critical production fixes (`hotfix/security-patch`)
- **docs/**: Documentation updates (`docs/api-reference`)
- **chore/**: Maintenance tasks (`chore/dependency-update`)

### Naming Rules
- Use lowercase with hyphens
- Be descriptive but concise
- Include issue number when applicable
- Examples:
  - `feature/subscription-filtering-#42`
  - `bugfix/delta-timeout-handling-#38`
  - `docs/websocket-api-reference`

## Git Workflow

### Development Workflow
1. **Create Feature Branch**:
   ```bash
   git checkout -b feature/subscription-filtering-#42
   ```

2. **Make Atomic Commits**:
   ```bash
   # Each commit should represent one logical change
   git add src/handlers/websocket_handler.go
   git commit -m "feat(websocket): add subscription state management"
   
   git add internal/handlers/websocket_handler_test.go
   git commit -m "test(websocket): add subscription management tests"
   ```

3. **Keep Branch Updated**:
   ```bash
   git checkout main
   git pull origin main
   git checkout feature/subscription-filtering-#42
   git rebase main  # Prefer rebase over merge for cleaner history
   ```

4. **Create Pull Request** with:
   - Clear title and description
   - Link to relevant issues
   - Test results and verification steps
   - Screenshots/examples if applicable

### Commit Size Guidelines
- **Atomic commits**: Each commit should represent one logical change
- **Small commits**: Easier to review and revert if needed
- **Complete commits**: Should not break the build or tests
- **Descriptive commits**: Should tell a story of development progress

## Pull Request Standards

### PR Title Format
Follow the conventional commit format:
```
type(scope): brief description of changes
```

Examples:
- `feat(websocket): implement client subscription management`
- `fix(delta-client): resolve connection timeout issues`
- `docs(api): update WebSocket message documentation`

### PR Description Template
```markdown
## Description
Brief summary of what this PR accomplishes.

## Changes Made
- [ ] Feature/fix 1 with brief description
- [ ] Feature/fix 2 with brief description
- [ ] Documentation updates

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Bug Fixes
If this PR fixes bugs, list them:
- Fixes #42: Description of bug fix
- Closes #38: Description of another fix

## Breaking Changes
If applicable, describe any breaking changes and migration steps.

## Verification Steps
1. Step to verify the changes work
2. Another verification step
3. Final verification step

## Screenshots/Examples
If applicable, add screenshots or code examples.
```

### PR Review Checklist
**For Authors**:
- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] Documentation updated if needed
- [ ] No merge conflicts
- [ ] Commit messages follow conventions
- [ ] PR description is complete

**For Reviewers**:
- [ ] Code logic is sound
- [ ] Test coverage is adequate
- [ ] Documentation is accurate
- [ ] Performance implications considered
- [ ] Security implications considered
- [ ] Breaking changes are documented

## Tag and Release Management

### Version Tagging
Use semantic versioning (semver):
- **Major**: Breaking changes (v2.0.0)
- **Minor**: New features, backward compatible (v1.1.0)
- **Patch**: Bug fixes, backward compatible (v1.0.1)

### Release Process
1. **Update Version Numbers**:
   ```bash
   # Update version in relevant files
   git add .
   git commit -m "chore: bump version to v1.2.0"
   ```

2. **Create and Push Tag**:
   ```bash
   git tag -a v1.2.0 -m "Release version 1.2.0"
   git push origin v1.2.0
   ```

3. **Create Release Notes**:
   ```markdown
   ## [1.2.0] - 2024-01-15
   
   ### Added
   - Client subscription management with filtering
   - Enhanced error handling for Delta Exchange connection
   
   ### Fixed
   - Connection timeout issues causing service crashes
   - Memory leak in WebSocket connection cleanup
   
   ### Changed
   - Improved logging with structured output
   - Updated API documentation with examples
   ```

## Bug Tracking Integration

### Issue References
Always reference issues in commits and PRs:
- **Fixes #123**: Closes the issue when merged
- **Closes #123**: Same as Fixes
- **Resolves #123**: Same as Fixes
- **Refs #123**: References but doesn't close
- **See #123**: References but doesn't close

### Bug Fix Workflow
1. **Create Issue**: Document the bug with reproduction steps
2. **Create Branch**: `bugfix/issue-description-#123`
3. **Fix and Test**: Implement fix with tests
4. **Document Fix**: Update bug report with solution
5. **Create PR**: Reference the issue in PR description
6. **Verify Fix**: Ensure issue is resolved before merging

## File and Directory Management

### Files to Always Track
- Source code files (*.go)
- Configuration files (*.yaml, *.json)
- Documentation (*.md)
- Build scripts (Makefile, Dockerfile)
- Test files (*_test.go)

### Files to Never Track (.gitignore)
```gitignore
# Binaries
main
websocket-service
*.exe

# Go specific
*.test
*.prof
coverage.out
coverage.html

# Environment files
.env
.env.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
*.tmp

# Build artifacts
build/
dist/
bin/
```

## Code Review Practices

### Before Requesting Review
- [ ] Self-review your changes
- [ ] Run all tests locally
- [ ] Check for merge conflicts
- [ ] Update documentation if needed
- [ ] Ensure commit messages are clear

### Review Guidelines
- **Be constructive**: Provide helpful feedback
- **Be specific**: Point to exact lines and suggest improvements
- **Ask questions**: If something is unclear, ask for clarification
- **Approve when ready**: Don't hold up good changes unnecessarily

### Common Review Comments
```
# Suggestion for improvement
Consider using a more descriptive variable name here.

# Question for clarification  
Why did you choose this approach over using a channel?

# Required change
This error handling is missing - could cause a panic.

# Nitpick (optional improvement)
nit: Consider adding a comment explaining this complex logic.

# Praise (encourage good practices)
Nice use of the decorator pattern here!
```

## Repository Maintenance

### Regular Maintenance Tasks
- **Weekly**: Review and update dependencies
- **Monthly**: Clean up merged branches
- **Quarterly**: Review and update documentation
- **As needed**: Update .gitignore and templates

### Branch Cleanup
```bash
# Delete merged local branches
git branch --merged main | grep -v main | xargs -n 1 git branch -d

# Delete remote tracking branches for deleted remotes
git remote prune origin
```

### Repository Health Checks
- Monitor repository size
- Check for sensitive data accidentally committed
- Ensure CI/CD pipelines are functioning
- Review access permissions periodically

These version control practices ensure a clean, traceable, and professional development workflow that supports effective collaboration and maintenance. 