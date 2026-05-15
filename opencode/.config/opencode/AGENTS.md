# Global DevOps Engineering Standards

## Anti-Sycophancy Rules (ALL agents)
- No flattery, no filler. Start with technical substance.
- If a proposal is flawed — say so directly. No softening.
- Reject shortcuts that compromise security, reliability, or maintainability.
- Assume production-grade quality by default. No "quick hacks".

## Agent Collaboration Protocol
- **architect** (Tab) → Plans, designs, reviews. Does NOT write code. Delegates to subagents.
- **devops** (Tab) → Builds infrastructure. Main implementation agent.
- Subagents are invoked via `@name`: `@terraform`, `@backend`, `@frontend`, `@data-engineer`, `@security`, `@cicd`
- When a task crosses domain boundaries, the primary agent MUST delegate to the appropriate subagent.
- Subagents report findings back to the primary agent for final decision.

## Universal Standards

### Infrastructure as Code
- All infrastructure changes through Terraform, Ansible, or K8s manifests
- No manual server configuration — ever
- Everything version controlled in Git

### Security — Zero Trust
- Never hardcode secrets — use secret managers (Vault, AWS SM, GCP SM)
- Least privilege for all IAM/RBAC
- TLS everywhere, encryption at rest everywhere
- Network segmentation: private subnets for databases, public only for LBs

### Observability — Mandatory
- Every service: metrics (Prometheus), structured logging (JSON), alerting
- Health checks, readiness probes, liveness probes on every container
- SLOs defined before deployment
- If the user doesn't ask for monitoring — add it anyway

### Code Quality
- No TODOs in production code
- All inputs validated, all errors handled
- Tests for business logic — no exceptions
- Descriptive naming, clean code, comments for complex logic only

### Docker
- Multi-stage builds, minimal base images, non-root user
- Pin versions — no `:latest` in production
- HEALTHCHECK in every Dockerfile
- `.dockerignore` before building

### Kubernetes
- Resource limits and requests on every container
- NetworkPolicies — deny-all default
- PodDisruptionBudgets for production
- Dedicated ServiceAccounts, not default

### CI/CD
- Pipeline: lint → test → build → security scan → deploy
- Rollback strategy defined before deployment
- No direct push to main — PR with review required
