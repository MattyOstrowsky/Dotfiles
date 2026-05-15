#!/usr/bin/env bash
# =============================================================================
# DevOps Workstation — Dependency Installer
# =============================================================================
# Usage: ./install.sh [--all | --core | --k8s | --iac | --dev | --security | --shell | --tui]
# Run with sudo where needed. Script is idempotent.
set -euo pipefail

YELLOW='\033[1;33m'
GREEN='\033[1;32m'
RED='\033[1;31m'
BLUE='\033[1;34m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[SKIP]${NC} $1 — already installed"; }
err()   { echo -e "${RED}[ERR]${NC} $1"; }

check_cmd() { command -v "$1" &>/dev/null; }

# =============================================================================
# Core CLI Tools
# =============================================================================
install_core() {
    info "=== Core CLI Tools ==="

    # fzf — fuzzy finder
    if check_cmd fzf; then warn "fzf"; else
        sudo apt install -y fzf && ok "fzf"
    fi

    # bat — better cat
    if check_cmd batcat || check_cmd bat; then warn "bat"; else
        sudo apt install -y bat && ok "bat"
    fi

    # jq — JSON processor
    if check_cmd jq; then warn "jq"; else
        sudo apt install -y jq && ok "jq"
    fi

    # yq — YAML processor
    if check_cmd yq; then warn "yq"; else
        sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
        sudo chmod +x /usr/local/bin/yq && ok "yq"
    fi

    # ripgrep — better grep
    if check_cmd rg; then warn "ripgrep"; else
        sudo apt install -y ripgrep && ok "ripgrep"
    fi

    # fd — better find
    if check_cmd fd || check_cmd fdfind; then warn "fd-find"; else
        sudo apt install -y fd-find && ok "fd-find"
    fi

    # tree
    if check_cmd tree; then warn "tree"; else
        sudo apt install -y tree && ok "tree"
    fi

    # htop
    if check_cmd htop; then warn "htop"; else
        sudo apt install -y htop && ok "htop"
    fi

    # direnv — auto-load .envrc
    if check_cmd direnv; then warn "direnv"; else
        sudo apt install -y direnv && ok "direnv"
    fi

    # zoxide — better cd
    if check_cmd zoxide; then warn "zoxide"; else
        curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh
        ok "zoxide"
    fi

    # tldr — simplified man pages
    if check_cmd tldr; then warn "tldr"; else
        pip3 install --user tldr && ok "tldr"
    fi
}

# =============================================================================
# Shell — Fish + Starship
# =============================================================================
install_shell() {
    info "=== Shell ==="

    # fish
    if check_cmd fish; then warn "fish"; else
        sudo apt-add-repository -y ppa:fish-shell/release-3
        sudo apt update && sudo apt install -y fish
        ok "fish"
    fi

    # starship prompt
    if check_cmd starship; then warn "starship"; else
        curl -sS https://starship.rs/install.sh | sh -s -- -y
        ok "starship"
    fi

    # fisher (fish plugin manager) — install inside fish
    info "Fisher plugins: run 'fisher update' inside fish after stow"
}

# =============================================================================
# Kubernetes Tools
# =============================================================================
install_k8s() {
    info "=== Kubernetes Tools ==="

    # kubectl
    if check_cmd kubectl; then warn "kubectl"; else
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
        rm -f kubectl && ok "kubectl"
    fi

    # helm
    if check_cmd helm; then warn "helm"; else
        curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        ok "helm"
    fi

    # kubectx + kubens — fast context/namespace switching
    if check_cmd kubectx; then warn "kubectx"; else
        sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
        sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
        sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens
        # Fish completions
        mkdir -p ~/.config/fish/completions
        ln -sf /opt/kubectx/completion/kubectx.fish ~/.config/fish/completions/
        ln -sf /opt/kubectx/completion/kubens.fish ~/.config/fish/completions/
        ok "kubectx + kubens"
    fi

    # k9s — terminal K8s UI
    if check_cmd k9s; then warn "k9s"; else
        curl -sSfL https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz | sudo tar xz -C /usr/local/bin k9s
        ok "k9s"
    fi

    # kustomize
    if check_cmd kustomize; then warn "kustomize"; else
        curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
        sudo mv kustomize /usr/local/bin/ && ok "kustomize"
    fi

    # stern — multi-pod log tailing
    if check_cmd stern; then warn "stern"; else
        go install github.com/stern/stern@latest && ok "stern"
    fi

    # kubeconform — K8s manifest validation
    if check_cmd kubeconform; then warn "kubeconform"; else
        go install github.com/yannh/kubeconform/cmd/kubeconform@latest && ok "kubeconform"
    fi
}

# =============================================================================
# Infrastructure as Code
# =============================================================================
install_iac() {
    info "=== Infrastructure as Code ==="

    # terraform
    if check_cmd terraform; then warn "terraform"; else
        sudo apt install -y gnupg software-properties-common
        wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
        echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
        sudo apt update && sudo apt install -y terraform
        ok "terraform"
    fi

    # terragrunt
    if check_cmd terragrunt; then warn "terragrunt"; else
        TGVER=$(curl -s https://api.github.com/repos/gruntwork-io/terragrunt/releases/latest | jq -r .tag_name)
        sudo wget -qO /usr/local/bin/terragrunt "https://github.com/gruntwork-io/terragrunt/releases/download/${TGVER}/terragrunt_linux_amd64"
        sudo chmod +x /usr/local/bin/terragrunt && ok "terragrunt"
    fi

    # ansible
    if check_cmd ansible; then warn "ansible"; else
        sudo apt install -y ansible && ok "ansible"
    fi

    # tfsec — Terraform static security analysis
    if check_cmd tfsec; then warn "tfsec"; else
        go install github.com/aquasecurity/tfsec/cmd/tfsec@latest && ok "tfsec"
    fi

    # checkov — IaC security scanner
    if check_cmd checkov; then warn "checkov"; else
        pip3 install --user checkov && ok "checkov"
    fi

    # infracost — cloud cost estimation
    if check_cmd infracost; then warn "infracost"; else
        curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
        ok "infracost"
    fi
}

# =============================================================================
# Development Languages & Tools
# =============================================================================
install_dev() {
    info "=== Development Tools ==="

    # go
    if check_cmd go; then warn "go"; else
        info "Install Go from https://go.dev/dl/"
    fi

    # python3 + pip
    if check_cmd python3; then warn "python3"; else
        sudo apt install -y python3 python3-pip python3-venv && ok "python3"
    fi

    # bun — JS/TS runtime & package manager
    if check_cmd bun; then warn "bun"; else
        curl -fsSL https://bun.sh/install | bash && ok "bun"
    fi

    # git
    if check_cmd git; then warn "git"; else
        sudo apt install -y git && ok "git"
    fi

    # neovim
    if check_cmd nvim; then warn "nvim"; else
        sudo apt install -y neovim && ok "nvim"
    fi
}

# =============================================================================
# TUI Tools — Terminal User Interfaces
# =============================================================================
install_tui() {
    info "=== TUI Tools ==="

    # lazygit — terminal git UI
    if check_cmd lazygit; then warn "lazygit"; else
        go install github.com/jesseduffield/lazygit@latest && ok "lazygit"
    fi

    # lazydocker — terminal docker UI
    if check_cmd lazydocker; then warn "lazydocker"; else
        go install github.com/jesseduffield/lazydocker@latest && ok "lazydocker"
    fi

    # lazysql — terminal database UI (Postgres, MySQL, SQLite)
    if check_cmd lazysql; then warn "lazysql"; else
        go install github.com/jorgerojas26/lazysql@latest && ok "lazysql"
    fi

    # dive — Docker image layer explorer
    if check_cmd dive; then warn "dive"; else
        DIVEVER=$(curl -s https://api.github.com/repos/wagoodman/dive/releases/latest | jq -r .tag_name)
        wget -qO /tmp/dive.deb "https://github.com/wagoodman/dive/releases/download/${DIVEVER}/dive_${DIVEVER#v}_linux_amd64.deb"
        sudo dpkg -i /tmp/dive.deb && rm /tmp/dive.deb && ok "dive"
    fi

    # btop — system monitor (better htop)
    if check_cmd btop; then warn "btop"; else
        sudo apt install -y btop && ok "btop"
    fi

    # ctop — container metrics TUI
    if check_cmd ctop; then warn "ctop"; else
        sudo wget -qO /usr/local/bin/ctop https://github.com/bcicen/ctop/releases/latest/download/ctop-linux-amd64
        sudo chmod +x /usr/local/bin/ctop && ok "ctop"
    fi

    # glow — markdown renderer in terminal
    if check_cmd glow; then warn "glow"; else
        go install github.com/charmbracelet/glow@latest && ok "glow"
    fi

    # navi — interactive cheatsheet (Ctrl+G)
    if check_cmd navi; then warn "navi"; else
        cargo install --locked navi 2>/dev/null || {
            BIN_DIR=/usr/local/bin
            curl -sL https://raw.githubusercontent.com/denisidoro/navi/master/scripts/install | bash
        }
        ok "navi"
    fi

    # atuin — better shell history with sync
    if check_cmd atuin; then warn "atuin"; else
        curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | sh
        ok "atuin"
    fi
}

# =============================================================================
# Security Tools
# =============================================================================
install_security() {
    info "=== Security Tools ==="

    # trivy — container/IaC vulnerability scanner
    if check_cmd trivy; then warn "trivy"; else
        sudo apt-get install -y wget apt-transport-https gnupg
        wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null
        echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
        sudo apt-get update && sudo apt-get install -y trivy
        ok "trivy"
    fi

    # grype — vulnerability scanner
    if check_cmd grype; then warn "grype"; else
        curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
        ok "grype"
    fi
}

# =============================================================================
# Main
# =============================================================================
show_help() {
    echo "Usage: ./install.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --all        Install everything"
    echo "  --core       Core CLI tools (fzf, bat, jq, yq, ripgrep, fd, direnv, zoxide)"
    echo "  --shell      Fish shell + Starship prompt"
    echo "  --k8s        Kubernetes tools (kubectl, helm, kubectx, k9s, stern, kustomize)"
    echo "  --iac        IaC tools (terraform, terragrunt, ansible, tfsec, checkov, infracost)"
    echo "  --dev        Dev tools (go, python, bun, git, nvim)"
    echo "  --tui        TUI tools (lazygit, lazydocker, lazysql, dive, btop, ctop, glow, navi, atuin)"
    echo "  --security   Security scanners (trivy, grype)"
    echo "  --list       List all tools and their install status"
    echo ""
    echo "Examples:"
    echo "  ./install.sh --all"
    echo "  ./install.sh --k8s --tui"
}

list_tools() {
    echo "=== Tool Status ==="
    echo "--- Core ---"
    for cmd in fish starship fzf batcat jq yq rg fdfind tree htop direnv zoxide tldr; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
    echo "--- Kubernetes ---"
    for cmd in kubectl helm kubectx kubens k9s kustomize stern kubeconform; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
    echo "--- IaC ---"
    for cmd in terraform terragrunt ansible tfsec checkov infracost; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
    echo "--- Dev ---"
    for cmd in go python3 bun git nvim; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
    echo "--- TUI ---"
    for cmd in lazygit lazydocker lazysql dive btop ctop glow navi atuin; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
    echo "--- Security ---"
    for cmd in trivy grype; do
        if check_cmd "$cmd"; then echo -e "  ${GREEN}✓${NC} $cmd"; else echo -e "  ${RED}✗${NC} $cmd"; fi
    done
}

if [[ $# -eq 0 ]]; then
    show_help
    exit 0
fi

for arg in "$@"; do
    case $arg in
        --all)      install_core; install_shell; install_k8s; install_iac; install_dev; install_tui; install_security ;;
        --core)     install_core ;;
        --shell)    install_shell ;;
        --k8s)      install_k8s ;;
        --iac)      install_iac ;;
        --dev)      install_dev ;;
        --tui)      install_tui ;;
        --security) install_security ;;
        --list)     list_tools ;;
        --help|-h)  show_help ;;
        *)          err "Unknown option: $arg"; show_help; exit 1 ;;
    esac
done

echo ""
ok "Done! Restart your shell or run 'exec fish' to apply changes."
