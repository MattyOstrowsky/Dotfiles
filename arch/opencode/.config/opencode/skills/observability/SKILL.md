---
name: observability
description: Monitoring, logging, and alerting patterns with Prometheus, Grafana, structured logging, SLO definition, and incident alerting setup
---
## Monitoring Stack Architecture

### Prometheus + Grafana
```
┌──────────┐     ┌────────────┐     ┌──────────┐
│  App      │────▶│ Prometheus │────▶│ Grafana  │
│ /metrics  │     │  (scrape)  │     │(dashboards)│
└──────────┘     └────────────┘     └──────────┘
                       │
                       ▼
                ┌──────────────┐
                │ Alertmanager │──▶ PagerDuty/Slack/Email
                └──────────────┘
```

### Key Metrics (RED Method)
- **R**ate — requests per second
- **E**rror — error rate (4xx, 5xx)
- **D**uration — request latency (P50, P95, P99)

### Key Metrics (USE Method) — for infrastructure
- **U**tilization — % resource busy
- **S**aturation — queue depth, pending work
- **E**rrors — error count

## Prometheus Configuration

### Service Scrape Config
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'app'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: (.+)
        replacement: $1
```

### Essential Recording Rules
```yaml
groups:
  - name: slo_rules
    rules:
    - record: job:http_request_duration_seconds:p99
      expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (job, le))
    - record: job:http_requests:error_rate
      expr: sum(rate(http_requests_total{status=~"5.."}[5m])) by (job) / sum(rate(http_requests_total[5m])) by (job)
    - record: job:http_requests:rate5m
      expr: sum(rate(http_requests_total[5m])) by (job)
```

## Alerting Rules

### Critical Alerts
```yaml
groups:
  - name: critical
    rules:
    - alert: HighErrorRate
      expr: job:http_requests:error_rate > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Error rate > 5% for {{ $labels.job }}"
        runbook: "https://wiki/runbooks/high-error-rate"

    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Pod {{ $labels.pod }} is crash-looping"

    - alert: DiskSpaceLow
      expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Disk space < 10% on {{ $labels.instance }}"

    - alert: HighMemoryUsage
      expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) > 0.9
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Memory usage > 90% on {{ $labels.instance }}"
```

### Alertmanager Routing
```yaml
# alertmanager.yml
route:
  receiver: default
  group_by: [alertname, namespace]
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - match:
        severity: critical
      receiver: pagerduty
      repeat_interval: 1h
    - match:
        severity: warning
      receiver: slack

receivers:
  - name: pagerduty
    pagerduty_configs:
      - service_key: '<key>'
  - name: slack
    slack_configs:
      - channel: '#alerts'
        send_resolved: true
  - name: default
    email_configs:
      - to: 'ops@example.com'
```

## SLO Definition Template

### Availability SLO
```
SLO: 99.9% availability (43.8 min downtime/month)
SLI: 1 - (count of 5xx responses / total responses)
Error budget: 0.1% of total requests per month
Alert: When error budget burn rate > 2x normal for 1 hour
```

### Latency SLO
```
SLO: P99 latency < 500ms
SLI: histogram_quantile(0.99, http_request_duration_seconds_bucket)
Alert: When P99 > 500ms for 5 minutes
```

## Structured Logging

### Log Format (JSON)
```json
{
  "timestamp": "2024-01-15T10:30:00.000Z",
  "level": "error",
  "service": "api-gateway",
  "trace_id": "abc123",
  "span_id": "def456",
  "message": "Request failed",
  "error": "connection timeout",
  "http_method": "POST",
  "http_path": "/api/v1/users",
  "http_status": 503,
  "duration_ms": 5002,
  "user_id": "usr_xxx"
}
```

### Log Levels
- **ERROR:** System failures, requires action
- **WARN:** Degraded but functional, investigate soon
- **INFO:** Significant business events (user created, payment processed)
- **DEBUG:** Diagnostic detail, never in production

### Rules
- Every log line must have: timestamp, level, service, trace_id
- No PII in logs (emails, passwords, tokens)
- Structured JSON, not free-text
- Correlation IDs across services for distributed tracing

## Dashboard Essentials (Grafana)
Every service dashboard must include:
1. **Request rate** — total QPS, by endpoint
2. **Error rate** — 4xx and 5xx, by endpoint
3. **Latency** — P50, P95, P99
4. **Saturation** — CPU, memory, disk, connections
5. **SLO burn rate** — remaining error budget
6. **Uptime** — current availability percentage
