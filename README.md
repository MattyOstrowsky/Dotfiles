# Dotfiles — DevOps Workstation

Personal dotfiles managed with [GNU Stow](https://www.gnu.org/software/stow/).
Includes an interactive dependency installer and 40+ navi cheatsheets.

## What's inside

| Package | Contents | Target |
|---------|----------|--------|
| `opencode/` | OpenCode AI agent config — 13 agents, 11 commands, 10 skills | `~/.config/opencode/` |
| `fish/` | Fish shell config — aliases, functions, plugins for DevOps | `~/.config/fish/` |
| `starship/` | Starship prompt — K8s, Terraform, Docker context | `~/.config/starship.toml` |
| `nvim/` | Neovim + NvChad config | `~/.config/nvim/` |
| `navi/` | Interactive cheatsheets (40+ tools) | `~/.config/navi/` |
| `lazygit/` | Git TUI — config with Nord theme | `~/.config/lazygit/` |
| `lazydocker/` | Docker TUI — config with Nord palette | `~/.config/lazydocker/` |
| `k9s/` | Kubernetes TUI — config, hotkeys, Nord skin | `~/.config/k9s/` |
| `btop/` | System monitor — config + Nord theme | `~/.config/btop/` |
| `bat/` | Better cat — Nord syntax theme | `~/.config/bat/` |
| `glow/` | Markdown renderer | `~/.config/glow/` |
| `atuin/` | Shell history with sync | `~/.config/atuin/` |
| `direnv/` | Per-directory environment | `~/.config/direnv/` |
| `dive/` | Docker image layer explorer | `~/.config/dive/` |

## Quick Start

### 1. Install prerequisites

```bash
sudo apt install stow git curl
```

### 2. Clone and symlink dotfiles

```bash
git clone https://github.com/MattyOstrowsky/Dotfiles.git ~/Dotfiles
cd ~/Dotfiles
make install
```

### 3. Install dependencies (interactive)

```bash
./install-deps.sh
```

Select categories you want to install. Already-installed tools are auto-skipped.
After each installation, matching configs are symlinked via stow automatically.

### Install specific package

```bash
cd ~/Dotfiles
stow lazygit    # Symlinks lazygit config → ~/.config/lazygit/
```

### Uninstall

```bash
cd ~/Dotfiles
stow -D lazygit   # Removes symlinks
make uninstall     # Uninstall all
```

### Dry run

```bash
make dry-run
```

## How GNU Stow works

Stow mirrors directory structure. Each top-level directory is a "package":

```
~/Dotfiles/
├── opencode    →  ~/.config/opencode/
├── fish        →  ~/.config/fish/
├── starship    →  ~/.config/starship.toml
├── nvim        →  ~/.config/nvim/
├── navi        →  ~/.config/navi/
├── lazygit     →  ~/.config/lazygit/
├── lazydocker  →  ~/.config/lazydocker/
├── k9s         →  ~/.config/k9s/
├── btop        →  ~/.config/btop/
├── bat         →  ~/.config/bat/
├── glow        →  ~/.config/glow/
├── atuin       →  ~/.config/atuin/
├── direnv      →  ~/.config/direnv/
└── dive        →  ~/.config/dive/
```

## Interactive Dependency Installer

`install-deps.sh` installs 50+ DevOps tools grouped by category:

| Category | Tools |
|----------|-------|
| **System** | stow, git, curl, wget, make, fish, fzf, bat, ripgrep, fd-find, tree, htop, direnv, btop, xclip |
| **Runtimes** | Python, Go, Rust, Node.js (via fnm), Bun |
| **Shell** | starship, zoxide, atuin, navi, fisher |
| **Editors** | neovim, opencode |
| **K8s** | kubectl, helm, k9s, kubectx, stern, kustomize, kubeconform |
| **IaC** | terraform, terragrunt, ansible, infracost |
| **TUI** | lazygit, lazydocker, lazysql, dive, ctop, glow |
| **Security** | trivy, grype, checkov, tldr |

Features:
- Numbered menu (default) or `--fzf` for fuzzy picker
- Auto-skips already-installed tools
- `--install` for non-interactive full install
- Auto-stows config after installation if a stow package exists
- All `go install` binaries land in `~/.local/bin/`

## Navi Cheatsheets

40+ interactive cheatsheets for daily tools. Launch with `navi` or query directly:

```bash
navi --query "helm install"
navi --query "lazygit keybindings"
```

See `navi/.config/navi/cheats/` for the full list.

## OpenCode Agents

### Primary Agents (switch with Tab)
- **daily** — Personal companion, planning, delegation (po polsku)
- **architect** — Architecture planning, ADRs, red-teaming (read-only)
- **orchestrator** — Complex task breakdown, execution plans (read-only)
- **devops** — Infrastructure implementation, Docker, K8s, CI/CD
- **meta** — Agent ecosystem management

### Subagents (invoke with @name)
`@terraform` `@ansible` `@backend` `@frontend` `@data-engineer` `@security` `@cicd` `@python-dev` `@explore`

### Commands
`tf-plan` `tf-apply` `docker-build` `k8s-check` `sec-audit` `pipeline-lint` `infra-review` `cost-estimate` `self-improve` `stats` `context-check`

## Fish Shell

### Key Aliases

| Alias | Command | Category |
|-------|---------|----------|
| `k` | `kubectl` | K8s |
| `kgp` | `kubectl get pods` | K8s |
| `k9` | `k9s` | K8s |
| `ksl` | `stern` | K8s |
| `kctx` | `kubectx` | K8s |
| `kns` | `kubens` | K8s |
| `tf` | `terraform` | Terraform |
| `tfp` | `terraform plan` | Terraform |
| `d` | `docker` | Docker |
| `dc` | `docker compose` | Docker |
| `lzg` | `lazygit` | TUI |
| `lzd` | `lazydocker` | TUI |
| `lzs` | `lazysql` | TUI |
| `top` | `btop` | TUI |
| `md` | `glow` | TUI |
| `ddive` | `dive` | TUI |
| `dtop` | `ctop` | TUI |
| `gs` | `git status` | Git |
| `gcp` | `git add -A && commit && push` | Git |
