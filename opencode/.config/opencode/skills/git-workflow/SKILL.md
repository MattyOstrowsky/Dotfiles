---
name: git-workflow
description: Git best practices, branch workflow, commit hygiene, PR review standards, merge strategies for DevOps workflows
---

## Branch Strategy

### Recommended: Trunk-Based Development
- Short-lived feature branches (max 1-2 days)
- Branch from `main`, merge back to `main`
- No long-running feature branches — decompose into small increments

### Naming Convention
```
{type}/{description}
feat/add-oidc-auth
fix/connection-pool-leak
chore/upgrade-terraform-provider
docs/update-readme
refactor/extract-monitoring-module
```

## Commit Hygiene

### Commit Message Format (Conventional Commits)
```
<type>(<scope>): <short summary>

[optional body]
[optional footer]
```

### Types
- `feat:` — New feature (triggers minor version bump)
- `fix:` — Bug fix (triggers patch version bump)
- `chore:` — Maintenance (no version bump)
- `refactor:` — Code restructuring with no behavior change
- `docs:` — Documentation only
- `test:` — Adding or updating tests
- `ci:` — CI/CD pipeline changes
- `sec:` — Security fix

### Commit Rules
- One logical change per commit — no "fix stuff" mega-commits
- Short summary: < 72 chars, capitalized, no trailing period
- Body: explain WHAT and WHY, not HOW
- Reference issues: `Fixes #123`, `Closes #456`
- No "WIP" commits in main branch — rebase before merge

```bash
# Good
git commit -m "fix(auth): handle token refresh race condition

The refresh token callback could fire multiple times when requests
are made in parallel. Added mutex lock around the refresh flow.

Fixes #234"
```

```bash
# Bad — reject these
git commit -m "fix stuff"
git commit -m "wip"
git commit -m "Update file"
```

## Pull Request Standards

### PR Checklist
- [ ] Tests pass locally (`pytest`, `npm test`, etc.)
- [ ] Linter passes
- [ ] No debugging artifacts (console.log, print, breakpoints)
- [ ] No TODOs or FIXMEs in production code
- [ ] Documentation updated if API/behavior changed
- [ ] Secrets checked in? (run `git diff --check`)

### PR Size
- Max 400 lines changed per PR
- If larger, split into multiple PRs
- One concern per PR — don't mix refactoring with feature work

### Review Flow
1. Author requests review
2. Reviewer checks: correctness, security, performance, naming
3. Author addresses feedback with new commits
4. Squash merge or rebase merge — no merge commits on main

## Merge Strategies

### Preferred: Squash Merge
```bash
# All commits in feature branch become one commit on main
git merge --squash feature-branch
# Benefits: clean main history, easy to revert
```

### Alternative: Rebase Merge
```bash
# Linear history, no merge commits
git checkout main && git rebase feature-branch
# Benefits: preserves atomic commits
```

### Reject: Merge Commits
```bash
# Creates ugly bubbles in history
git merge --no-ff feature-branch
# Avoid unless there's a specific reason for the merge commit
```

## Git Operations Safety

### Never Do
```bash
git push --force                        # Unless you own the branch
git push --force-with-lease             # Slightly safer but still dangerous on shared branches
git reset --hard HEAD~1                 # If already pushed
git commit --amend                      # If already pushed and shared
git rebase main                         # If branch is shared with others
git add -A && git commit               # Without reviewing changes first
```

### Always Do
```bash
git status                              # Review before committing
git diff --staged                       # Review staged changes
git log --oneline --graph              # Understand branch state
git pull --rebase                       # Instead of git pull (creates merge commit)
```

## Tagging & Releases

### Semantic Versioning
```
v{major}.{minor}.{patch}
v1.2.3  # Public API
v1.2.3-rc.1  # Release candidate
v1.2.3-alpha.1  # Alpha
```

### Tag Creation
```bash
git tag -a v1.2.3 -m "Release v1.2.3: Added OIDC auth, fixed connection pool leak"
git push origin v1.2.3
```

## Anti-Patterns
- ❌ Committing directly to main (requires branch protection)
- ❌ Large PRs (>400 lines)
- ❌ No PR description — reviewer shouldn't guess the intent
- ❌ Merge commits on main (prefer squash or rebase)
- ❌ Force pushing to shared branches
- ❌ .env, secrets, or credentials in repository
- ❌ Binary files in repository (use Git LFS or object storage)
