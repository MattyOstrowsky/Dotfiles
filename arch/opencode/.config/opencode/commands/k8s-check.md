---
description: Validate Kubernetes manifests and check best practices
agent: devops
---
Analyze Kubernetes manifests in the current directory (or $ARGUMENTS path).

Steps:
1. **Syntax validation:** If `kubectl` is available, run `kubectl apply --dry-run=client -f $ARGUMENTS`

2. **Best practices check:** For each manifest, verify:
   - Resource limits and requests defined
   - Liveness and readiness probes configured
   - Security context (non-root, read-only filesystem)
   - Image tags pinned (no `latest`)
   - Labels and annotations present (app, version, team)
   - Namespace specified (not default)
   - PodDisruptionBudget for Deployments
   - NetworkPolicy exists
   - ServiceAccount specified (not default)

3. **Report:** Generate a summary with:
   - ✅ Passing checks
   - ⚠️ Warnings (best practice violations)
   - ❌ Critical issues (security problems)

4. **If `kubeconform` or `kubeval` available:** Run schema validation against the target K8s version
