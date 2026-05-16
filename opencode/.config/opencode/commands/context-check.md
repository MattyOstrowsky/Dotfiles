---
description: Verify the current request is in-scope for the active agent, or delegate to the correct subagent
agent: devops
subtask: true
---
Analyze the user's request in $ARGUMENTS and determine if it's within your agent's scope.

## Scope Classification

Determine which category the request falls into:

| Category | In Scope? | Action |
|----------|-----------|--------|
| Infrastructure as Code (Terraform, Ansible, K8s) | ✅ Yes | Proceed |
| Docker/Containerization | ✅ Yes | Proceed |
| CI/CD pipeline design | ✅ Yes | Proceed |
| Monitoring/Observability | ✅ Yes | Proceed |
| Security hardening | ✅ Yes | Use @security |
| Cloud cost optimization | ✅ Yes | Use @data-engineer |
| Python/Shell automation scripts | ✅ Yes | Use @python-dev |
| Backend API code (Node.js, Go, Python) | ⚠️ Partial | Use @backend |
| Frontend/UI code | ❌ No | Use @frontend |
| Mobile app development | ❌ No | Flag as CONTEXT MISMATCH |
| Game development | ❌ No | Flag as CONTEXT MISMATCH |
| Machine Learning model training | ❌ No | Flag as CONTEXT MISMATCH |

## Output

If in scope:
```
SCOPE CHECK: This request is within [agent role] scope.
Delegating to: [subagent if needed, or "self"]
```

If out of scope:
```
CONTEXT MISMATCH: This request is outside my [agent role] scope.
This appears to be: [topic]
Recommended action: [delegate to X | ask user to clarify | suggest different tool]
```
