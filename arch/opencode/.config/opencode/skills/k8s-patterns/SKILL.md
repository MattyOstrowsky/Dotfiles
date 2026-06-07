---
name: k8s-patterns
description: Kubernetes deployment patterns, resource configuration, scaling strategies, and production-ready manifest templates
---
## Deployment Patterns

### Blue-Green Deployment
```yaml
# Deploy v2 alongside v1, switch Service selector when ready
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-green
  labels:
    app: myapp
    version: v2
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
      version: v2
  template:
    metadata:
      labels:
        app: myapp
        version: v2
---
# Switch traffic by updating selector
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: myapp
    version: v2  # Flip from v1 to v2
```

### Canary Deployment
- Deploy canary with small replica count
- Use Istio VirtualService or Nginx weight-based routing
- Monitor error rates for 15min before full rollout
- Automated rollback on error rate > threshold

### Rolling Update (default)
```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # Max pods over desired count
      maxUnavailable: 0    # Zero downtime
```

## Production-Ready Pod Template
```yaml
spec:
  containers:
  - name: app
    image: myapp:1.2.3  # Pinned version, never :latest
    ports:
    - containerPort: 8080
      protocol: TCP
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 15
      periodSeconds: 10
      failureThreshold: 3
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
    startupProbe:        # Slow-starting containers
      httpGet:
        path: /healthz
        port: 8080
      failureThreshold: 30
      periodSeconds: 10
    securityContext:
      runAsNonRoot: true
      runAsUser: 1000
      readOnlyRootFilesystem: true
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
    volumeMounts:
    - name: tmp
      mountPath: /tmp
  volumes:
  - name: tmp
    emptyDir: {}
  serviceAccountName: app-sa  # Dedicated SA, not default
  automountServiceAccountToken: false
```

## Scaling

### HPA (Horizontal Pod Autoscaler)
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300  # Wait 5min before scaling down
```

### PodDisruptionBudget
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: app
spec:
  minAvailable: 2  # Or maxUnavailable: 1
  selector:
    matchLabels:
      app: myapp
```

## Network Security
```yaml
# Deny all by default
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# Allow specific traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-app
spec:
  podSelector:
    matchLabels:
      app: myapp
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: database
    ports:
    - port: 5432
```

## Anti-Patterns
- ❌ No resource limits → noisy neighbor, OOM kills
- ❌ No probes → dead pods receiving traffic
- ❌ `latest` tag → unpredictable deployments
- ❌ Default namespace → no isolation
- ❌ Default ServiceAccount → over-privileged pods
- ❌ No PDB → all pods killed during node drain
- ❌ No NetworkPolicy → unrestricted lateral movement
