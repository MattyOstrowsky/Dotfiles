# Dotfiles — DevOps Workstation

Personal dotfiles managed with [GNU Stow](https://www.gnu.org/software/stow/).

## What's inside

| Package | Contents | Target |
|---------|----------|--------|
| `opencode/` | OpenCode AI agent config — 8 agents, 9 commands, 7 skills | `~/.config/opencode/` |

## Quick Start

### Prerequisites

```bash
sudo apt install stow    # Debian/Ubuntu
# or
brew install stow         # macOS
```

### Install all dotfiles

```bash
git clone https://github.com/MattyOstrowsky/Dotfiles.git ~/Dotfiles
cd ~/Dotfiles
make install
```

### Install specific package

```bash
cd ~/Dotfiles
stow opencode    # Symlinks opencode/.config/opencode/* → ~/.config/opencode/*
```

### Uninstall

```bash
cd ~/Dotfiles
stow -D opencode   # Removes symlinks, leaves original files untouched
# or
make uninstall      # Uninstall all
```

### Check what would happen (dry run)

```bash
stow -n -v opencode   # Shows what symlinks would be created without doing it
```

## How GNU Stow works

Stow mirrors directory structure. Each top-level directory is a "package":

```
~/Dotfiles/
└── opencode/                          ← Stow package
    └── .config/
        └── opencode/
            ├── opencode.json          → ~/.config/opencode/opencode.json
            ├── AGENTS.md              → ~/.config/opencode/AGENTS.md
            ├── tui.json               → ~/.config/opencode/tui.json
            ├── agents/
            │   ├── architect.md       → ~/.config/opencode/agents/architect.md
            │   ├── devops.md          → ...
            │   ├── terraform.md
            │   ├── backend.md
            │   ├── frontend.md
            │   ├── data-engineer.md
            │   ├── security.md
            │   └── cicd.md
            ├── commands/
            │   ├── tf-plan.md
            │   ├── tf-apply.md
            │   ├── docker-build.md
            │   ├── k8s-check.md
            │   ├── sec-audit.md
            │   ├── pipeline-lint.md
            │   ├── infra-review.md
            │   ├── cost-estimate.md
            │   └── self-improve.md
            └── skills/
                ├── terraform-debug/SKILL.md
                ├── docker-best-practices/SKILL.md
                ├── k8s-patterns/SKILL.md
                ├── cicd-patterns/SKILL.md
                ├── cloud-cost/SKILL.md
                ├── security-checklist/SKILL.md
                └── observability/SKILL.md
```

## Adding new packages

To add a new config (e.g. `fish` shell):

```bash
# 1. Create stow structure
mkdir -p ~/Dotfiles/fish/.config/fish

# 2. Move your existing config
mv ~/.config/fish/config.fish ~/Dotfiles/fish/.config/fish/

# 3. Create symlink
cd ~/Dotfiles
stow fish

# 4. Commit
git add fish/
git commit -m "feat: add fish config"
```

## OpenCode Agents Overview

### Primary Agents (switch with Tab)
- **architect** — Architecture planning, ADRs, red-teaming (read-only)
- **devops** — Infrastructure implementation, Docker, K8s, CI/CD (full access)

### Subagents (invoke with @name)
- `@terraform` — Terraform IaC specialist
- `@backend` — API, databases, server-side code
- `@frontend` — UI, dashboards, web tooling
- `@data-engineer` — ETL, data modeling, pipelines
- `@security` — Security audits, hardening, STRIDE
- `@cicd` — CI/CD pipelines, GitHub Actions, GitLab CI

### Slash Commands
| Command | Description |
|---------|-------------|
| `/tf-plan` | Terraform validate + plan with impact analysis |
| `/tf-apply` | Apply reviewed Terraform plan |
| `/docker-build` | Lint, build, analyze Docker images |
| `/k8s-check` | Validate K8s manifests + best practices |
| `/sec-audit` | Security audit on infrastructure code |
| `/pipeline-lint` | Lint CI/CD pipeline configurations |
| `/infra-review` | Architecture review with ADR format |
| `/cost-estimate` | Cloud cost estimation |
| `/self-improve` | Analyze session and improve agent configs |
