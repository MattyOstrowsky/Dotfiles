---
description: Apply a reviewed Terraform plan
agent: terraform
---
First, check if `plan.tfplan` exists in the current directory (or in $ARGUMENTS directory if provided).

If the plan file exists:
1. Run `terraform show plan.tfplan` and display a summary of changes
2. Highlight any DESTROY operations with a clear warning
3. Show the total count of resources: to create, to update, to destroy
4. Ask for confirmation before running `terraform apply plan.tfplan`

If no plan file exists:
1. Inform the user to run `/tf-plan` first
2. Do NOT run `terraform apply` without a saved plan
