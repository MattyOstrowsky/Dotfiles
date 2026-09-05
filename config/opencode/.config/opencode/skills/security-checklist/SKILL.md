---
name: security-checklist
description: Infrastructure security hardening checklist covering network, IAM, secrets, containers, Kubernetes, databases, and compliance frameworks
---
## Network Security

### Firewall / Security Groups
- [ ] Default deny-all inbound and outbound
- [ ] Allow only specific ports from specific sources
- [ ] No 0.0.0.0/0 on any port except 80/443 for public-facing services
- [ ] Separate security groups per service tier (web, app, db)
- [ ] No SSH/RDP open to the internet — use bastion or SSM/IAP

### Network Architecture
- [ ] Public subnet: only load balancers and bastion hosts
- [ ] Private subnet: application servers, databases
- [ ] Isolated subnet: sensitive data processing
- [ ] VPC peering/PrivateLink for cross-account communication
- [ ] DNS over HTTPS/TLS for internal resolution
- [ ] WAF in front of all public endpoints

### TLS/SSL
- [ ] TLS 1.2+ everywhere — no exceptions
- [ ] HSTS headers with long max-age
- [ ] Certificate auto-renewal (cert-manager, ACM)
- [ ] Internal mTLS between services (Istio, Linkerd)

## Identity & Access Management

### IAM Policies
- [ ] Least privilege — no `*` in actions or resources
- [ ] No root/admin account for daily use
- [ ] MFA enforced on all human accounts
- [ ] Service accounts with minimal scoped roles
- [ ] Regular access reviews (quarterly minimum)
- [ ] Unused credentials rotated or deleted after 90 days

### RBAC (Kubernetes)
- [ ] No ClusterAdmin for workloads
- [ ] Namespace-scoped roles where possible
- [ ] Dedicated ServiceAccount per workload
- [ ] `automountServiceAccountToken: false` unless needed
- [ ] Audit RBAC with `kubectl auth can-i --list`

## Secrets Management

### Storage
- [ ] All secrets in a secret manager (Vault, AWS Secrets Manager, GCP Secret Manager)
- [ ] No secrets in: source code, CI/CD logs, environment variables, ConfigMaps
- [ ] Rotation policy defined for every secret
- [ ] Sealed Secrets or External Secrets Operator for K8s
- [ ] `.env` files in `.gitignore` — always

### Encryption
- [ ] Encryption at rest for all databases and storage
- [ ] KMS-managed keys, not default encryption
- [ ] Separate keys per environment (dev/staging/prod)
- [ ] Key rotation enabled (annual minimum)

## Container Security

### Image Security
- [ ] Base images from trusted registries only
- [ ] Images scanned for CVEs before deployment (Trivy, Snyk)
- [ ] No `latest` tag — pin specific versions
- [ ] Image signing and verification (Cosign, Notary)
- [ ] Minimal base images (distroless, alpine, scratch)

### Runtime Security
- [ ] Containers run as non-root
- [ ] Read-only root filesystem
- [ ] No privileged mode
- [ ] Capabilities dropped: `ALL`
- [ ] Seccomp/AppArmor profiles applied
- [ ] No host namespace sharing (PID, network, IPC)

## Database Security
- [ ] No public access — private subnet only
- [ ] SSL/TLS connections enforced
- [ ] Credentials rotated regularly
- [ ] Audit logging enabled
- [ ] Backups encrypted and tested for restoration
- [ ] Query logging for sensitive tables
- [ ] Parameterized queries only — no SQL string concatenation

## Logging & Monitoring
- [ ] Centralized logging (CloudWatch, Stackdriver, ELK)
- [ ] Audit logs for all cloud API calls (CloudTrail, Audit Logs)
- [ ] VPC Flow Logs enabled
- [ ] Alerting on: auth failures, privilege escalation, unusual traffic
- [ ] Log retention: 90 days hot, 1 year cold (compliance dependent)
- [ ] No sensitive data in logs (PII, tokens, passwords)

## Compliance Frameworks Reference
| Framework | Focus | Key Requirements |
|-----------|-------|-----------------|
| SOC 2 | Security/Availability | Access controls, monitoring, encryption |
| ISO 27001 | Info Security | Risk management, ISMS, audit trail |
| PCI DSS | Payment Data | Network segmentation, encryption, access control |
| HIPAA | Health Data | PHI encryption, access logs, BAA |
| GDPR | Personal Data | Data minimization, right to erasure, consent |
| CIS Benchmarks | Hardening | Specific config rules per cloud/OS |

## Incident Response
1. **Detect:** Automated alerting on anomalies
2. **Contain:** Isolate affected systems (revoke creds, block IPs)
3. **Eradicate:** Remove root cause
4. **Recover:** Restore from clean backups
5. **Post-mortem:** Blameless review, update runbooks
