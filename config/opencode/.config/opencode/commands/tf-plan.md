---
description: Validate and plan Terraform changes
agent: terraform
---
Run `terraform validate` on all .tf files in the current directory. If validation passes, run `terraform plan -out=plan.tfplan` and analyze the output.

For each resource change:
1. Identify if it's CREATE, UPDATE, or DESTROY
2. Flag any DESTROY operations with a warning
3. Check for potential issues (missing tags, public access, no encryption)
4. Summarize the total impact

If $ARGUMENTS is provided, run the plan in that directory: `cd $ARGUMENTS && terraform validate && terraform plan -out=plan.tfplan`
