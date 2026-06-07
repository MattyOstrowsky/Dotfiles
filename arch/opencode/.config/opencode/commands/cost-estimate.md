---
description: Estimate cloud costs for infrastructure or queries
agent: data-engineer
subtask: true
---
Analyze the infrastructure code or query in $ARGUMENTS and estimate cloud costs.

## Analysis Targets

### Terraform Resources
- Parse `.tf` files and identify billable resources
- Estimate monthly cost based on instance types, storage, data transfer
- Compare with cheaper alternatives (reserved instances, spot, preemptible)
- Flag over-provisioned resources

### Database Queries (BigQuery/Snowflake/Redshift)
- Estimate data scanned (bytes processed)
- Calculate query cost based on provider pricing
- Suggest optimizations: partitioning, clustering, materialized views
- Compare full scan vs incremental approach

### Kubernetes Workloads
- Calculate resource requests × node pricing
- Identify over-provisioned pods (requests >> actual usage)
- Suggest right-sizing based on resource limits

## Output Format
```
| Resource              | Type          | Monthly Est. | Optimization        | Savings  |
|-----------------------|---------------|-------------|---------------------|----------|
| RDS db.r5.xlarge      | Database      | $XXX        | Reserved Instance   | ~40%     |
| S3 bucket (1TB)       | Storage       | $XX         | Lifecycle rules     | ~30%     |
| EKS 3x m5.large       | Compute       | $XXX        | Spot instances      | ~60%     |
```

Total estimated monthly: $XXX
Potential savings: $XXX (~XX%)
