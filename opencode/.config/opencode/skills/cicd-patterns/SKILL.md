---
name: cicd-patterns
description: CI/CD pipeline patterns and templates for GitHub Actions, GitLab CI, and ArgoCD including caching, security scanning, and deployment strategies
---
## GitHub Actions Patterns

### Optimized Workflow Template
```yaml
name: CI/CD
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

permissions:
  contents: read
  id-token: write  # OIDC

jobs:
  lint:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: 22
        cache: npm
    - run: npm ci
    - run: npm run lint

  test:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: lint
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: 22
        cache: npm
    - run: npm ci
    - run: npm test -- --coverage
    - uses: actions/upload-artifact@v4
      with:
        name: coverage
        path: coverage/

  security:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: lint
    steps:
    - uses: actions/checkout@v4
    - name: Dependency scan
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: fs
        severity: CRITICAL,HIGH

  build:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [test, security]
    steps:
    - uses: actions/checkout@v4
    - name: Configure AWS OIDC
      uses: aws-actions/configure-aws-credentials@v4
      with:
        role-to-assume: ${{ vars.AWS_ROLE_ARN }}
        aws-region: eu-central-1
    - name: Build & push
      run: |
        docker build -t $ECR/$IMAGE:${{ github.sha }} .
        docker push $ECR/$IMAGE:${{ github.sha }}

  deploy-staging:
    needs: build
    environment: staging
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to staging
      run: kubectl set image deployment/app app=$ECR/$IMAGE:${{ github.sha }}

  deploy-production:
    needs: deploy-staging
    environment:
      name: production
      url: https://app.example.com
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to production
      run: kubectl set image deployment/app app=$ECR/$IMAGE:${{ github.sha }}
```

### Reusable Composite Action
```yaml
# .github/actions/setup/action.yml
name: Setup Project
description: Install deps and cache
runs:
  using: composite
  steps:
  - uses: actions/setup-node@v4
    with:
      node-version: 22
      cache: npm
  - run: npm ci
    shell: bash
```

## GitLab CI Patterns

### Optimized Pipeline
```yaml
stages:
  - lint
  - test
  - build
  - security
  - deploy

variables:
  DOCKER_BUILDKIT: "1"

.node-cache: &node-cache
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - node_modules/
    policy: pull

lint:
  stage: lint
  <<: *node-cache
  script:
    - npm ci
    - npm run lint
  cache:
    policy: pull-push  # First job populates cache

test:
  stage: test
  <<: *node-cache
  script:
    - npm ci
    - npm test -- --coverage
  coverage: '/Statements\s*:\s*(\d+\.?\d*%)/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml

build:
  stage: build
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH

deploy-staging:
  stage: deploy
  environment:
    name: staging
  script:
    - kubectl set image deployment/app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH

deploy-production:
  stage: deploy
  environment:
    name: production
  script:
    - kubectl set image deployment/app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      when: manual
```

## ArgoCD GitOps Pattern

### Application Manifest
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/org/k8s-manifests
    targetRevision: main
    path: apps/myapp/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    retry:
      limit: 3
      backoff:
        duration: 5s
        maxDuration: 3m
```

## Pipeline Anti-Patterns
- ❌ No timeout → stuck jobs consume runners
- ❌ Unpinned action versions → supply chain attacks
- ❌ Secrets in logs → credential exposure
- ❌ No concurrency control → wasted resources
- ❌ Sequential when parallelizable → slow feedback
- ❌ No cache → slow builds, unnecessary downloads
- ❌ No approval gate for production → accidental deployments
