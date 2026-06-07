---
name: cloud-cost
description: Cloud cost optimization strategies for AWS, GCP, and Azure including compute, storage, networking, and database right-sizing
---
## Compute Cost Optimization

### AWS EC2 / GCP Compute / Azure VMs
- **Reserved Instances / Committed Use:** 1-3 year commitments for predictable workloads → 40-72% savings
- **Spot/Preemptible Instances:** For stateless, fault-tolerant workloads → up to 90% savings
- **Right-sizing:** Analyze CPU/memory utilization; downsize if avg < 40%
- **Savings Plans (AWS):** Flexible commitment across instance families
- **Auto-scaling:** Scale to zero in non-prod during off-hours

### Kubernetes
- **Right-size pods:** Set requests based on P95 usage, not peak
- **Cluster autoscaler:** Scale nodes based on pending pods
- **Spot node pools:** For non-critical workloads
- **Karpenter (AWS):** More efficient node provisioning than cluster autoscaler
- **Vertical Pod Autoscaler:** Auto-adjust requests based on actual usage

## Storage Optimization

### Object Storage (S3/GCS/Blob)
- **Lifecycle policies:** Transition to cheaper tiers automatically
  - Standard → Infrequent Access (30 days)
  - IA → Glacier/Archive (90 days)
  - Archive → Deep Archive (180 days)
  - Delete after retention period
- **Intelligent Tiering (AWS):** Auto-moves objects between tiers
- **Compression:** Gzip/Zstd before upload for text data
- **Cleanup:** Delete incomplete multipart uploads, old versions

### Database
- **Aurora Serverless / Cloud SQL autoscaling:** Pay per actual usage
- **Read replicas:** Cheaper than scaling primary
- **Reserved capacity:** For predictable IOPS/throughput
- **DynamoDB on-demand vs provisioned:** On-demand for spiky, provisioned for steady

## Networking
- **NAT Gateway costs (AWS):** ~$0.045/GB — use VPC endpoints for S3/DynamoDB
- **Data transfer:** Keep traffic within same AZ/region when possible
- **CloudFront/CDN:** Cache at edge, reduce origin traffic
- **VPC Endpoints:** Eliminate NAT Gateway costs for AWS service traffic

## Monitoring Cost
```bash
# AWS Cost Explorer CLI
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=SERVICE

# GCP BigQuery cost analysis
SELECT service.description, SUM(cost) as total
FROM `project.dataset.gcp_billing_export_v1_*`
WHERE DATE(usage_start_time) >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
GROUP BY 1 ORDER BY 2 DESC
```

## Cost Estimation Formulas
- **EC2:** (instance_price/hr × hours/month) + EBS + data_transfer
- **RDS:** instance_price + storage_price + IOPS + backup + data_transfer
- **S3:** storage_per_GB × volume + requests × price_per_1000 + data_transfer
- **K8s cluster:** (node_count × node_price) + LB + data_transfer + EBS

## Quick Wins Checklist
- [ ] Delete unattached EBS volumes / persistent disks
- [ ] Terminate stopped instances with no scheduled restart
- [ ] Remove unused Elastic IPs / static IPs
- [ ] Delete old snapshots beyond retention policy
- [ ] Review and delete unused load balancers
- [ ] Consolidate underutilized accounts/projects
- [ ] Set billing alerts at 50%, 80%, 100% of budget
- [ ] Tag everything for cost allocation
