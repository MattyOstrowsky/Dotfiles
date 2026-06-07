---
name: docker-best-practices
description: Docker and Docker Compose best practices for building secure, efficient, production-ready container images and multi-service stacks
---
## Dockerfile Best Practices

### Base Image Selection
- Use specific version tags: `node:22.5-slim` not `node:latest`
- Prefer minimal images: `alpine`, `slim`, `distroless`
- For security-critical: use `gcr.io/distroless/static` or `chainguard/static`
- Pin digest for maximum reproducibility: `FROM node:22@sha256:abc...`

### Multi-Stage Builds
```dockerfile
# Stage 1: Build
FROM node:22-slim AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --production=false
COPY . .
RUN npm run build

# Stage 2: Runtime
FROM gcr.io/distroless/nodejs22
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
USER nonroot
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s CMD ["node", "-e", "require('http').get('http://localhost:3000/health')"]
CMD ["dist/main.js"]
```

### Layer Optimization
1. Order layers from least to most frequently changing
2. Combine RUN commands with `&&` to reduce layers
3. Use `.dockerignore` to exclude: `.git`, `node_modules`, `*.md`, `tests/`
4. Clean up in the same RUN command: `apt-get install -y ... && rm -rf /var/lib/apt/lists/*`

### Security
- Never run as root: `USER nonroot` or `USER 1000:1000`
- No secrets in ENV or ARG — use BuildKit secrets: `--mount=type=secret`
- Scan images: `trivy image <tag>` or `docker scout cves <tag>`
- Set read-only filesystem where possible
- Drop all capabilities: `--cap-drop=ALL`

### Health Checks
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

## Docker Compose Best Practices

### Service Configuration
```yaml
services:
  app:
    image: myapp:1.2.3  # Pin versions
    build:
      context: .
      dockerfile: Dockerfile
      target: production  # Multi-stage target
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    read_only: true
    tmpfs:
      - /tmp
    security_opt:
      - no-new-privileges:true
```

### Networking
- Use custom networks, never `network_mode: host`
- Separate frontend/backend networks
- Internal networks for backend services

### Volumes & Data
- Named volumes for persistent data
- Bind mounts only for development
- Never store secrets in volumes — use Docker secrets or env files

### Environment Variables
- Use `.env` file for local dev, never commit it
- `env_file:` for service-specific config
- Required vars documented in `.env.example`
