# Dotfiles — DevOps Workstation

Personal dotfiles managed with [GNU Stow](https://www.gnu.org/software/stow/).

## What's inside

| Package | Contents | Target |
|---------|----------|--------|
| `opencode/` | OpenCode AI agent config — 12 agents, 11 commands, 10 skills | `~/.config/opencode/` |
| `pi/` | Pi AI agent config — agents, extensions, prompts | `~/.pi/agent/` |
| `fish/` | Fish shell config — aliases, functions, plugins for DevOps | `~/.config/fish/` |
| `starship/` | Starship prompt — K8s, Terraform, Docker, AWS context | `~/.config/starship.toml` |
| `nvim/` | Neovim + NvChad config | `~/.config/nvim/` |
| `navi/` | Interactive cheatsheets (docker, git, k8s, terraform) | `~/.config/navi/` |

## Quick Start

### Prerequisites

```bash
sudo apt install stow git curl    # Debian/Ubuntu
brew install stow git             # macOS
```

See [dependencies.md](dependencies.md) for a full dependency list.

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

### Dry run

```bash
make dry-run
```

## How GNU Stow works

Stow mirrors directory structure. Each top-level directory is a "package":

```
~/Dotfiles/
├── opencode/ → ~/.config/opencode/
├── pi/       → ~/.pi/agent/
├── fish/     → ~/.config/fish/
├── starship/ → ~/.config/starship.toml
├── nvim/     → ~/.config/nvim/
└── navi/     → ~/.config/navi/
```

## Adding new packages

```bash
mkdir -p ~/Dotfiles/mypkg/.config/myapp
mv ~/.config/myapp/config.toml ~/Dotfiles/mypkg/.config/myapp/
cd ~/Dotfiles && stow myapp
git add myapp/ && git commit -m "feat: add myapp config"
```

## OpenCode Agents

### Primary Agents (switch with Tab)
- **architect** — Architecture planning, ADRs, red-teaming (read-only)
- **orchestrator** — Complex task breakdown, execution plans, delegating to subagents
- **devops** — Infrastructure implementation, Docker, K8s, CI/CD (full access)
- **meta** — Agent ecosystem management, building new agents/skills/commands

### Subagents (invoke with @name)
`@terraform` `@ansible` `@backend` `@frontend` `@data-engineer` `@security` `@cicd` `@python-dev`

### Slash Commands
`/tf-plan` `/tf-apply` `/docker-build` `/k8s-check` `/sec-audit` `/pipeline-lint` `/infra-review` `/cost-estimate` `/self-improve` `/stats`

## Fish Shell

### Key Aliases

| Alias | Command | Category |
|-------|---------|----------|
| `k` | `kubectl` | K8s |
| `kgp` | `kubectl get pods` | K8s |
| `tf` | `terraform` | Terraform |
| `tfp` | `terraform plan` | Terraform |
| `d` | `docker` | Docker |
| `dc` | `docker compose` | Docker |
| `gs` | `git status` | Git |
| `gcp` | `git add -A && commit && push` | Git |
