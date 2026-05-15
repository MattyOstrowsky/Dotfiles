---
description: Build and analyze Docker images
agent: devops
---
Analyze the Dockerfile in the current directory (or $1 if provided) and then build it.

Steps:
1. **Lint first:** Check the Dockerfile for best practices:
   - Multi-stage build used?
   - Running as non-root?
   - HEALTHCHECK defined?
   - No `latest` base image tags?
   - `.dockerignore` exists?
   - Layers ordered for cache efficiency?

2. **Build:** Run `docker build -t $2 .` (or auto-generate tag from directory name if $2 not provided)

3. **Analyze:** After successful build:
   - Show image size
   - Run `docker history` to show layer sizes
   - Flag any layers over 100MB
   - Suggest optimizations if image is over 500MB

4. **Security:** If `trivy` is available, run `trivy image <tag>` for CVE scan
