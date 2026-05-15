---
description: Run security audit on infrastructure code
agent: security
---
Perform a security audit on the file or directory specified in $ARGUMENTS (or current directory if not provided).

## Audit Scope

### Terraform files (.tf)
- Scan for hardcoded secrets, access keys, passwords
- Check IAM policies for overly permissive rules (`*` actions/resources)
- Verify encryption at rest for storage and databases
- Check security groups for open ports (0.0.0.0/0)
- Verify no public access on S3 buckets, RDS instances
- Run `tfsec` or `checkov` if available

### Docker files (Dockerfile)
- Running as root?
- Using untrusted base images?
- Secrets in build args or ENV?
- COPY of sensitive files?

### Kubernetes manifests (.yaml/.yml)
- Privileged containers?
- Host network/PID namespace?
- Missing security contexts?
- Secrets in plain text?
- Missing network policies?

### CI/CD pipelines (.github/workflows/, .gitlab-ci.yml)
- Secrets exposure in logs
- Unpinned action versions
- Missing security scanning steps

## Output Format
For each finding:
- **[CRITICAL/HIGH/MEDIUM/LOW]** Severity
- **File:Line** — Location
- **Finding** — What's wrong
- **Impact** — What an attacker can exploit
- **Fix** — Exact remediation steps
