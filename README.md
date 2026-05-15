# Dotfiles — DevOps Workstation

Personal dotfiles managed with [GNU Stow](https://www.gnu.org/software/stow/).

## What's inside

| Package | Contents | Target |
|---------|----------|--------|
| `opencode/` | OpenCode AI agent config — 8 agents, 9 commands, 7 skills | `~/.config/opencode/` |
| `fish/` | Fish shell config — aliases, functions, plugins for DevOps | `~/.config/fish/` |
| `starship/` | Starship prompt — K8s, Terraform, Docker, AWS context in prompt | `~/.config/starship.toml` |

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

## Fish Shell Setup

### Prerequisites

```bash
sudo apt install fish fzf bat          # Core tools
curl -sS https://starship.rs/install.sh | sh  # Starship prompt

# Optional but recommended
sudo apt install direnv                 # Auto-load .envrc files
cargo install zoxide                    # Better cd (or: sudo apt install zoxide)
```

### After `stow fish && stow starship`

Install Fisher plugins:

```bash
# Install Fisher (plugin manager)
curl -sL https://raw.githubusercontent.com/jorgebucaran/fisher/main/functions/fisher.fish | source && fisher install jorgebucaran/fisher

# Install all plugins from fish_plugins
fisher update
```

### Fish Plugins

| Plugin | What it does |
|--------|-------------|
| [fisher](https://github.com/jorgebucaran/fisher) | Plugin manager for fish |
| [fzf.fish](https://github.com/PatrickF1/fzf.fish) | Fuzzy finder — `Ctrl+R` history, `Ctrl+F` files, `Ctrl+Alt+S` git status |
| [z](https://github.com/jethrokuan/z) | Jump to frequently used directories — `z dotfiles` |
| [autopair](https://github.com/jorgebucaran/autopair.fish) | Auto-close brackets, quotes, backticks |
| [sponge](https://github.com/meaningful-ooo/sponge) | Clean failed commands from history automatically |
| [abbreviation-tips](https://github.com/gazorby/fish-abbreviation-tips) | Shows hints when you type a command that has an abbreviation |
| [puffer-fish](https://github.com/nickeb96/puffer-fish) | `!!` (last command) and `!$` (last argument) expansion like bash |
| [replay](https://github.com/jorgebucaran/replay.fish) | Run bash commands in fish — `replay 'source .env'` |
| [pisces](https://github.com/laughedelic/pisces) | Paired symbols auto-insertion (brackets, quotes) |

### Key Aliases

| Alias | Command | Category |
|-------|---------|----------|
| `k` | `kubectl` | K8s |
| `kgp` | `kubectl get pods` | K8s |
| `kns` | `kubectl config set-context --current --namespace` | K8s |
| `tf` | `terraform` | Terraform |
| `tfp` | `terraform plan` | Terraform |
| `d` | `docker` | Docker |
| `dc` | `docker compose` | Docker |
| `gs` | `git status` | Git |
| `gcp` | `git add -A && commit && push` | Git |

### Starship Prompt Segments

The prompt shows contextual info only when relevant:

```
[OS] [user] [~/path] [git:branch ✓ +3-1] [⎈ k8s:ns] [💠 tf] [🐳 docker] [☁️ aws] [🐍 py3.12 (venv)] [🐹 go1.22]
❯ 
```

Segments: `purple` OS/user → `blue` path → `green` git → `orange` devops tools → `purple` languages → `red` duration/status


