---
name: terraform-debug
description: Debug and troubleshoot Terraform infrastructure issues including state drift, provider errors, dependency cycles, and plan failures
---
## Diagnostic Workflow

### Step 1: Syntax & Config Validation
```bash
terraform validate
terraform fmt -check -recursive
```

### Step 2: State Inspection
```bash
terraform state list                    # List all tracked resources
terraform state show <resource>         # Show resource details
terraform plan -refresh-only            # Detect drift without changes
terraform state pull | jq '.resources[] | .type + "." + .name'  # Quick state dump
```

### Step 3: Dependency Analysis
```bash
terraform graph | dot -Tsvg > graph.svg     # Full dependency graph
terraform graph -type=plan                   # Plan-specific dependencies
```

### Step 4: Verbose Debugging
```bash
TF_LOG=DEBUG terraform plan 2>debug.log     # Full debug output
TF_LOG=TRACE terraform plan 2>trace.log     # Maximum verbosity
TF_LOG_CORE=DEBUG terraform plan            # Core runtime only
TF_LOG_PROVIDER=DEBUG terraform plan        # Provider operations only
```

### Step 5: Console Testing
```bash
terraform console                            # Interactive expression testing
# > aws_vpc.main.id
# > cidrsubnet("10.0.0.0/16", 8, 1)
# > length(var.subnets)
```

## Common Issues & Fixes

### State Drift
- **Symptom:** Plan shows changes to resources you didn't modify
- **Cause:** Resource modified outside Terraform (manual change, other tool)
- **Fix:** `terraform plan -refresh-only` to review, then `terraform apply -refresh-only` to sync state
- **Prevention:** Lock down manual access, use `lifecycle { ignore_changes = [...] }` for expected drift

### Circular Dependencies
- **Symptom:** `Error: Cycle` in plan output
- **Cause:** Resources referencing each other directly
- **Fix:** Use `depends_on` explicitly, split into modules, or use data sources

### State Lock Issues
- **Symptom:** `Error: Error locking state`
- **Cause:** Previous run crashed, another user running
- **Fix:** `terraform force-unlock <LOCK_ID>` (verify no one else is running first)

### Provider Auth Failures
- **Symptom:** `Error: No valid credential sources found`
- **Fix:** Check `AWS_PROFILE`, `GOOGLE_APPLICATION_CREDENTIALS`, or az login status

### Missing Provider
- **Symptom:** `provider registry.terraform.io/hashicorp/xxx not found`
- **Fix:** `terraform init -upgrade`

### Resource Already Exists
- **Symptom:** `Error: already exists`
- **Fix:** `terraform import <resource_address> <real_resource_id>` to adopt existing resource

### State File Corruption
- **Symptom:** `Error: Unsupported state file format`
- **Fix:** Restore from versioned backup in S3, never manually edit state
