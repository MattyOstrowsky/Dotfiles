---
description: Lint and validate CI/CD pipeline configuration
agent: cicd
---
Analyze CI/CD pipeline configuration in $ARGUMENTS (or auto-detect from current directory).

## Auto-detection
1. Check for `.github/workflows/*.yml` — GitHub Actions
2. Check for `.gitlab-ci.yml` — GitLab CI
3. Check for `Jenkinsfile` — Jenkins
4. Check for `bitbucket-pipelines.yml` — Bitbucket

## Validation

### GitHub Actions
- Valid YAML syntax
- All referenced actions exist and versions are pinned (SHA preferred)
- `concurrency` configured to prevent duplicate runs
- Caching configured for dependencies
- OIDC authentication used (no long-lived secrets)
- Required workflow steps present: lint, test, build, security scan
- Timeout configured on each job
- Proper use of `permissions:` (least privilege)

### GitLab CI
- Valid YAML syntax (run `gitlab-ci-lint` if available)
- Stages defined and ordered correctly
- `rules:` used instead of deprecated `only:/except:`
- Cache and artifacts configured properly
- Protected environments for production
- `needs:` for DAG pipeline optimization

### General
- No secrets in plain text
- Rollback strategy defined
- Notification/alerting on failure
- Deploy steps have approval gates for production

## Output
- ✅ Valid configurations
- ⚠️ Best practice violations
- ❌ Critical issues (security, missing stages)
- 💡 Optimization suggestions (parallelism, caching)
