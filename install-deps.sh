#!/usr/bin/env bash
#
# Install-Deps — Interactive dependency installer for Dotfiles
#
# Usage:
#   ./install-deps.sh            # Interactive mode (recommended)
#   ./install-deps.sh --list     # List all tools with status
#   ./install-deps.sh --install  # Non-interactive, installs all missing
#
# Features:
#   - Category-based selection with fzf or fallback numbered menu
#   - Checks if each tool already exists before offering to install
#   - Shows preview of what will be installed, asks for confirmation
#   - Safe to re-run — skips already-installed tools

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$SCRIPT_DIR"
SELF="$0"

# Ensure common binary paths are in PATH for exists() checks
export PATH="$HOME/go/bin:$HOME/.local/bin:$HOME/.atuin/bin:$PATH"

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ─── Helper: check if a command exists ──────────────────────────────────────
exists() {
  # Special case: fish functions (like fisher) aren't PATH binaries
  if [[ "$1" == "fisher" ]]; then
    fish -c "type -q fisher" &>/dev/null
    return $?
  fi
  command -v "$1" &>/dev/null
}

# ─── Helper: check tool installation status ─────────────────────────────────
check() {
  if exists "$1"; then
    return 0
  fi
  return 1
}

status_icon() {
  if check "$1"; then
    echo -e "${GREEN}✓${NC}"
  else
    echo -e "${RED}✗${NC}"
  fi
}

status_label() {
  if check "$1"; then
    echo -e "${GREEN}[installed]${NC}"
  else
    echo -e "${RED}[missing]${NC}"
  fi
}

# ─── Tool definitions ────────────────────────────────────────────────────────
# Each entry: category | name | binary | install_func | needs | via
#
# `needs` = runtime required: "go", "cargo", "pip", "bun", "fnm", "" (none)
# `via`   = display label for install method
#
# Categories are processed IN ORDER. Runtimes MUST be first so they
# get installed before tools that depend on them.
declare -a TOOLS=()

tool() {
  local category="$1"
  local name="$2"
  local binary="$3"
  local install_func="$4"
  local needs="${5:-}"
  local via="${6:-}"
  TOOLS+=("$category|$name|$binary|$install_func|$needs|$via")
}

# Helper to safely increment (bash set -e hates ((x++)) when x=0)
inc() { local -n v="$1"; v=$((v + 1)); }

# ─── Install functions ───────────────────────────────────────────────────────

# ── System / apt packages ──────────────────────────────────────────────────
sys_pkg() {
  local pkg="$1"
  if rpm -q "$pkg" &>/dev/null 2>&1; then return 0; fi
  sudo dnf install -y "$pkg" 2>&1 | tail -1
}
install_stow()     { sys_pkg stow; }
install_git()      { sys_pkg git; }
install_curl()     { sys_pkg curl; }
install_wget()     { sys_pkg wget; }
install_make()     { sys_pkg make; }
install_fish()     { sys_pkg fish; }
install_fzf()      { sys_pkg fzf; }
install_bat()      { sys_pkg bat; }
install_ripgrep()  { sys_pkg ripgrep; }
install_fdfind()   { sys_pkg fd-find; }
install_tree()     { sys_pkg tree; }
install_htop()     { sys_pkg htop; }
install_direnv()   { sys_pkg direnv; }
install_btop()     { sys_pkg btop; }
install_unzip()    { sys_pkg unzip; }
install_xclip()    { sys_pkg xclip; }

# ── Runtimes (no sudo — user space) ────────────────────────────────────────

install_python() {
  if exists python3; then
    echo "  $(python3 --version 2>/dev/null)"
    return 0
  fi
  sudo dnf install -y python3 python3-pip python3-venv 2>&1 | tail -1
}
# pip3 comes with python3 package, but make sure
install_pip3() {
  if exists pip3; then
    echo "  $(pip3 --version 2>/dev/null)"
    return 0
  fi
  sudo dnf install -y python3-pip 2>&1 | tail -1
}

install_node() {
  # Use fnm (Fast Node Manager) — no sudo, per-user node versions
  if ! exists fnm; then
    curl -fsSL https://fnm.vercel.app/install | bash 2>/dev/null
    # Source fnm for this session
    export PATH="$HOME/.local/share/fnm:$PATH"
    eval "$(fnm env --use-on-cd 2>/dev/null)"
  fi
  if exists node; then
    echo "  node $(node --version 2>/dev/null) + npm $(npm --version 2>/dev/null)"
    return 0
  fi
  # Install LTS node
  fnm install --lts 2>&1 | tail -1
  fnm use lts-latest 2>/dev/null
  echo "  node $(node --version 2>/dev/null) installed via fnm"
}

install_go() {
  if exists go; then
    echo "  $(go version)"
    return 0
  fi
  # Install to ~/.local/go (no sudo needed)
  wget -q https://go.dev/dl/go1.22.2.linux-amd64.tar.gz -O /tmp/go.tar.gz
  mkdir -p "$HOME/.local"
  rm -rf "$HOME/.local/go"
  tar -C "$HOME/.local" -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  echo "  Go installed to $HOME/.local/go"
  echo "  Add to PATH: export PATH=\"\$HOME/.local/go/bin:\$PATH\""
}

install_rust() {
  if exists cargo; then
    echo "  $(cargo --version)"
    return 0
  fi
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
}

install_bun() {
  if exists bun; then
    echo "  $(bun --version)"
    return 0
  fi
  curl -fsSL https://bun.sh/install | bash
}

# ── Shell tools ────────────────────────────────────────────────────────────

install_starship() {
  if exists starship; then return 0; fi
  curl -sS https://starship.rs/install.sh | sh -s -- -y
}

install_zoxide() {
  if exists zoxide; then return 0; fi
  curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh
}

install_atuin() {
  if exists atuin; then return 0; fi
  curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | sh
}

install_navi() {
  if exists navi; then return 0; fi
  if ! exists cargo; then echo "  ⚠ rust/cargo required"; return 1; fi
  cargo install --locked navi
}

install_fisher() {
  if ! exists fish; then echo "  ⚠ fish required"; return 1; fi
  fish -c 'curl -sL https://raw.githubusercontent.com/jorgebucaran/fisher/main/functions/fisher.fish | source && fisher install jorgebucaran/fisher && fisher update' 2>/dev/null
}

# ── Editors ────────────────────────────────────────────────────────────────

install_nvim() {
  if exists nvim; then
    echo "  $(nvim --version 2>/dev/null | head -1)"
    return 0
  fi
  curl -LO https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.appimage
  chmod +x nvim-linux-x86_64.appimage
  sudo mv nvim-linux-x86_64.appimage /usr/local/bin/nvim
}

install_opencode() {
  if exists opencode; then return 0; fi
  # opencode is installed via bun
  curl -fsSL https://opencode.ai/install | bash
}

# ── Kubernetes ─────────────────────────────────────────────────────────────

install_kubectl() {
  if exists kubectl; then
    echo "  $(kubectl version --client 2>/dev/null | head -1)"
    return 0
  fi
  local ver
  ver=$(curl -sL https://dl.k8s.io/release/stable.txt)
  curl -sLO "https://dl.k8s.io/release/$ver/bin/linux/amd64/kubectl"
  install -m 0755 kubectl "$HOME/.local/bin/kubectl"
  rm -f kubectl
}

install_helm() {
  if exists helm; then
    echo "  $(helm version --short 2>/dev/null)"
    return 0
  fi
  local ver
  ver=$(curl -s https://api.github.com/repos/helm/helm/releases/latest | grep '"tag_name"' | cut -d'"' -f4)
  wget -q "https://get.helm.sh/helm-${ver}-linux-amd64.tar.gz" -O /tmp/helm.tar.gz
  tar xzf /tmp/helm.tar.gz -C /tmp/ && install -m 755 /tmp/linux-amd64/helm "$HOME/.local/bin/helm"
  rm -rf /tmp/helm.tar.gz /tmp/linux-amd64
  echo "  Helm $ver installed to ~/.local/bin"
}

install_kubectx() {
  if exists kubectx && exists kubens; then return 0; fi
  sudo git clone --depth 1 https://github.com/ahmetb/kubectx /opt/kubectx 2>/dev/null || true
  sudo ln -sf /opt/kubectx/kubectx /usr/local/bin/kubectx
  sudo ln -sf /opt/kubectx/kubens /usr/local/bin/kubens
  mkdir -p "$HOME/.config/fish/completions"
  ln -sf /opt/kubectx/completion/kubectx.fish "$HOME/.config/fish/completions/" 2>/dev/null || true
  ln -sf /opt/kubectx/completion/kubens.fish "$HOME/.config/fish/completions/" 2>/dev/null || true
}

install_k9s() {
  if exists k9s; then return 0; fi
  curl -sSfL https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz | sudo tar xz -C /usr/local/bin k9s
}

install_stern() {
  if exists stern; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/stern/stern@latest
}

install_kustomize() {
  if exists kustomize; then return 0; fi
  curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
  sudo mv kustomize /usr/local/bin/ 2>/dev/null || true
}

install_kubeconform() {
  if exists kubeconform; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/yannh/kubeconform/cmd/kubeconform@latest
}

# ── IaC ────────────────────────────────────────────────────────────────────

install_terraform() {
  if exists terraform; then
    echo "  $(terraform --version 2>/dev/null | head -1)"
    return 0
  fi
  # Binary install — avoids repo issues on newer Fedora
  local ver
  ver=$(curl -s https://api.github.com/repos/hashicorp/terraform/releases/latest | grep '"tag_name"' | cut -d'"' -f4 | sed 's/^v//')
  wget -q "https://releases.hashicorp.com/terraform/${ver}/terraform_${ver}_linux_amd64.zip" -O /tmp/tf.zip
  unzip -qo /tmp/tf.zip -d /tmp/tf-install && sudo install -m 755 /tmp/tf-install/terraform /usr/local/bin/terraform
  rm -rf /tmp/tf.zip /tmp/tf-install
  echo "  Terraform $ver installed"
}

install_terragrunt() {
  if exists terragrunt; then return 0; fi
  local ver
  ver=$(curl -s https://api.github.com/repos/gruntwork-io/terragrunt/releases/latest | grep '"tag_name"' | cut -d'"' -f4)
  sudo wget -qO /usr/local/bin/terragrunt "https://github.com/gruntwork-io/terragrunt/releases/download/$ver/terragrunt_linux_amd64"
  sudo chmod +x /usr/local/bin/terragrunt
}

install_ansible() {
  if exists ansible; then return 0; fi
  sys_pkg ansible
}

install_infracost() {
  if exists infracost; then return 0; fi
  curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
}

# ── TUI ────────────────────────────────────────────────────────────────────

install_lazygit() {
  if exists lazygit; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/jesseduffield/lazygit@latest
}

install_lazydocker() {
  if exists lazydocker; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  # Pin v0.23.3 — newer versions have a Go 1.25 compatibility issue
  GOBIN="$HOME/.local/bin" go install github.com/jesseduffield/lazydocker@v0.23.3
}

install_lazysql() {
  if exists lazysql; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/jorgerojas26/lazysql@latest
}

install_dive() {
  if exists dive; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/wagoodman/dive@latest
}

install_ctop() {
  if exists ctop; then return 0; fi
  wget -qO "$HOME/.local/bin/ctop" https://github.com/bcicen/ctop/releases/latest/download/ctop-linux-amd64
  chmod +x "$HOME/.local/bin/ctop"
}

install_glow() {
  if exists glow; then return 0; fi
  if ! exists go; then echo "  ⚠ go required"; return 1; fi
  GOBIN="$HOME/.local/bin" go install github.com/charmbracelet/glow@latest
}

# ── Security ───────────────────────────────────────────────────────────────

install_trivy() {
  if exists trivy; then return 0; fi
  sudo dnf install -y trivy 2>/dev/null || {
    rpm -q trivy 2>/dev/null && return 0
    sudo dnf install -y wget 2>/dev/null
    wget -qO - https://aquasecurity.github.io/trivy-repo/rpm/public.key | sudo tee /etc/pki/rpm-gpg/RPM-GPG-KEY-TRIVY > /dev/null
    echo -e "[trivy]\nname=Trivy\nbaseurl=https://aquasecurity.github.io/trivy-repo/rpm/releases/\$releasever/\$basearch/\ngpgkey=https://aquasecurity.github.io/trivy-repo/rpm/public.key\nenabled=1\ngpgcheck=1" | sudo tee /etc/yum.repos.d/trivy.repo > /dev/null
    sudo dnf install -y trivy
  }
}

install_grype() {
  if exists grype; then return 0; fi
  curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
}

install_checkov() {
  if exists checkov; then return 0; fi
  pip3 install --user --break-system-packages checkov 2>&1 | tail -1
}

install_tldr() {
  if exists tldr; then return 0; fi
  # Use --prefix because system npm needs root for default prefix (/usr)
  npm install -g tldr --prefix="$HOME/.local" 2>&1 | tail -3
  # Add ~/.local/bin to PATH for this session
  hash -r 2>/dev/null || true
}

# ─── Auto-stow: symlink config from Dotfiles after install ─────────────────────
# Maps tool_name → stow package directory name
STOW_PACKAGES=(
  "lazygit:lazygit"
  "lazydocker:lazydocker"
  "k9s:k9s"
  "btop:btop"
  "bat:bat"
  "glow:glow"
  "atuin:atuin"
  "direnv:direnv"
  "dive:dive"
  "navi:navi"
  "starship:starship"
  "fish:fish"
  "nvim:neovim"
)

auto_stow() {
  local binary="$1"
  local tool_name="$2"

  # Map tool to stow package name
  local pkg=""
  for mapping in "${STOW_PACKAGES[@]}"; do
    local key="${mapping%%:*}"
    local val="${mapping#*:}"
    if [[ "$key" == "$binary" ]] || [[ "$key" == "$tool_name" ]]; then
      pkg="$val"
      break
    fi
  done

  # If no mapping, try using the tool name directly
  [[ -z "$pkg" ]] && pkg="$tool_name"

  local pkg_dir="$DOTFILES_DIR/$pkg"
  if [[ -d "$pkg_dir" ]]; then
    echo -e "  ${BLUE}→${NC} stowing config: ${CYAN}$pkg${NC}"
    stow --target="$HOME" "$pkg" 2>/dev/null && echo -e "    ${GREEN}✓${NC} config linked" || echo -e "    ${YELLOW}⚠${NC} stow failed (maybe already linked)"
  fi
}

# ─── Register all tools ──────────────────────────────────────────────────────
# tool "category" "name" "binary" "install_func" "needs" "via"
#
# Runtimes MUST be first category — tools that need go/cargo/pip/fnm are
# installed in order and will pick up the runtime if it was selected.

# ── System (dnf packages only) ─────────────────────────────────────────────
tool "System"     "stow"      "stow"      install_stow     ""    "dnf"
tool "System"     "git"       "git"       install_git      ""    "dnf"
tool "System"     "curl"      "curl"      install_curl     ""    "dnf"
tool "System"     "wget"      "wget"      install_wget     ""    "dnf"
tool "System"     "make"      "make"      install_make     ""    "dnf"
tool "System"     "unzip"     "unzip"     install_unzip    ""    "dnf"
tool "System"     "fish"      "fish"      install_fish     ""    "dnf"
tool "System"     "fzf"       "fzf"       install_fzf      ""    "dnf"
tool "System"     "bat"       "bat"       install_bat      ""    "dnf"
tool "System"     "ripgrep"   "rg"        install_ripgrep  ""    "dnf"
tool "System"     "fd-find"   "fdfind"    install_fdfind   ""    "dnf"
tool "System"     "tree"      "tree"      install_tree     ""    "dnf"
tool "System"     "htop"      "htop"      install_htop     ""    "dnf"
tool "System"     "direnv"    "direnv"    install_direnv   ""    "dnf"
tool "System"     "btop"      "btop"      install_btop     ""    "dnf"
tool "System"     "xclip"     "xclip"     install_xclip    ""    "dnf"

# ── Runtimes (language runtimes + pkg managers, no sudo) ────────────────────
tool "Runtimes"   "Python"    "python3"   install_python   ""    "dnf"
tool "Runtimes"   "pip3"      "pip3"      install_pip3     ""    "dnf"
tool "Runtimes"   "Node.js"   "node"      install_node     ""    "fnm"
tool "Runtimes"   "Go"        "go"        install_go       ""    "tarball"
tool "Runtimes"   "Rust"      "cargo"     install_rust     ""    "rustup"
tool "Runtimes"   "Bun"       "bun"       install_bun      ""    "curl"

# ── Shell enhancements ─────────────────────────────────────────────────────
tool "Shell"     "starship"  "starship"  install_starship ""    "curl"
tool "Shell"     "zoxide"    "zoxide"    install_zoxide   ""    "curl"
tool "Shell"     "atuin"     "atuin"     install_atuin    ""    "curl"
tool "Shell"     "navi"      "navi"      install_navi     "cargo"  "cargo install"
tool "Shell"     "fisher"    "fisher"    install_fisher   ""    "fish"

# ── Editors ────────────────────────────────────────────────────────────────
tool "Editors"   "neovim"    "nvim"      install_nvim     ""    "appimage"
tool "Editors"   "opencode"  "opencode"  install_opencode "bun" "bun install"

# ── Kubernetes ─────────────────────────────────────────────────────────────
tool "K8s"       "kubectl"   "kubectl"   install_kubectl  ""    "binary"
tool "K8s"       "helm"      "helm"      install_helm     ""    "script"
tool "K8s"       "kubectx"   "kubectx"   install_kubectx  ""    "git clone"
tool "K8s"       "k9s"       "k9s"       install_k9s      ""    "binary"
tool "K8s"       "stern"     "stern"     install_stern    "go"  "go install"
tool "K8s"       "kustomize" "kustomize" install_kustomize ""   "script"
tool "K8s"       "kubeconform" "kubeconform" install_kubeconform "go" "go install"

# ── Infrastructure as Code ─────────────────────────────────────────────────
tool "IaC"       "terraform" "terraform" install_terraform ""   "dnf"
tool "IaC"       "terragrunt" "terragrunt" install_terragrunt "" "binary"
tool "IaC"       "ansible"   "ansible"   install_ansible  ""    "dnf"
tool "IaC"       "infracost" "infracost" install_infracost ""  "script"

# ── Terminal UI tools ──────────────────────────────────────────────────────
tool "TUI"       "lazygit"   "lazygit"   install_lazygit  "go"  "go install"
tool "TUI"       "lazydocker" "lazydocker" install_lazydocker "go" "go install"
tool "TUI"       "lazysql"   "lazysql"   install_lazysql  "go"  "go install"
tool "TUI"       "dive"      "dive"      install_dive     "go"  "go install"
tool "TUI"       "ctop"      "ctop"      install_ctop     ""    "binary"
tool "TUI"       "glow"      "glow"      install_glow     "go"  "go install"

# ── Security ───────────────────────────────────────────────────────────────
tool "Security"  "trivy"     "trivy"     install_trivy    ""    "dnf"
tool "Security"  "grype"     "grype"     install_grype    ""    "script"
tool "Security"  "checkov"   "checkov"   install_checkov  "pip" "pip3 install"
tool "Security"  "tldr"      "tldr"      install_tldr     "fnm" "npm i -g"

# ─── Functions ───────────────────────────────────────────────────────────────

# Get unique categories in order
get_categories() {
  for entry in "${TOOLS[@]}"; do
    echo "${entry%%|*}"
  done | awk '!seen[$0]++'  # unique, keep order
}

# List tools in a category (indices into TOOLS array)
tools_in_category() {
  local cat="$1"
  local i=0
  for entry in "${TOOLS[@]}"; do
    if [[ "${entry%%|*}" == "$cat" ]]; then
      echo "$i"
    fi
    i=$((i + 1))
  done
}

# Get field from tool entry
# 1=category, 2=name, 3=binary, 4=install_func, 5=needs, 6=via
tool_field() {
  local idx="$1"
  local field="$2"
  IFS='|' read -r -a parts <<< "${TOOLS[$idx]}"
  echo "${parts[$((field-1))]}"
}

# Parse a tool entry into named vars: sets $cat $name $bin $func $needs $via
parse_tool() {
  local entry="$1"
  IFS='|' read -r cat name bin func needs via <<< "$entry"
}

# Human label for what a tool needs
needs_label() {
  local needs="$1"
  case "$needs" in
    go)    echo "needs Go" ;;
    cargo) echo "needs Rust" ;;
    pip)   echo "needs pip3" ;;
    bun)   echo "needs Bun" ;;
    fnm)   echo "needs Node.js" ;;
    *)     echo "" ;;
  esac
}

# ─── Display: list all tools with status ────────────────────────────────────
list_all() {
  local prev_cat=""
  for entry in "${TOOLS[@]}"; do
    IFS='|' read -r cat name binary func needs via <<< "$entry"
    if [[ "$cat" != "$prev_cat" ]]; then
      echo -e "\n${BOLD}${CYAN}── $cat ──${NC}"
      prev_cat="$cat"
    fi
    local icon
    icon=$(status_icon "$binary")
    printf "  %s %-12s %b  %b\n" "$icon" "$name" "$(status_label "$binary")" "${YELLOW}[$via]${NC}"
  done
}

# ─── Display: dry-run (show what would be installed) ────────────────────────
show_selection() {
  local -n sel_ref="$1"
  local count=0
  for idx in "${!sel_ref[@]}"; do
    if [[ "${sel_ref[$idx]}" == "1" ]]; then
      count=$((count + 1))
    fi
  done

  if [[ "$count" -eq 0 ]]; then
    echo -e "\n${YELLOW}Nothing selected. Exiting.${NC}"
    exit 0
  fi

  echo -e "\n${BOLD}${BLUE}Selected for installation (${count} tools):${NC}"
  local prev_cat=""
  for idx in "${!sel_ref[@]}"; do
    [[ "${sel_ref[$idx]}" != "1" ]] && continue
    IFS='|' read -r cat name binary func needs via <<< "${TOOLS[$idx]}"
    if [[ "$cat" != "$prev_cat" ]]; then
      echo -e "\n  ${CYAN}$cat:${NC}"
      prev_cat="$cat"
    fi
    local icon
    icon=$(status_icon "$binary")
    echo -e "    $icon $name  ${YELLOW}[$via]${NC}"
  done

  echo ""
  read -r -p "Proceed with installation? [Y/n] " confirm
  case "$confirm" in
    n|N|no|NO) echo -e "${YELLOW}Aborted.${NC}"; exit 0 ;;
    *) echo -e "${GREEN}Installing...${NC}" ;;
  esac
}

# ─── Install selected tools (grouped by category, runtimes first) ─────────
install_selected() {
  local -n sel_ref="$1"
  local fail_count=0
  local skip_count=0
  local ok_count=0

  # Map: need → binary to check
  need_check() {
    case "$1" in
      go)    echo "go" ;;
      cargo) echo "cargo" ;;
      pip)   echo "pip3" ;;
      bun)   echo "bun" ;;
      fnm)   echo "node" ;;
      *)     echo "" ;;
    esac
  }

  # Collect categories in order
  local cats=()
  while IFS= read -r c; do cats+=("$c"); done < <(get_categories)

  for c in "${cats[@]}"; do
    # Check if anything selected in this category
    local has=0
    for idx in "${!sel_ref[@]}"; do
      [[ "${sel_ref[$idx]}" != "1" ]] && continue
      IFS='|' read -r cat name binary func needs via <<< "${TOOLS[$idx]}"
      [[ "$cat" == "$c" ]] && { has=1; break; }
    done
    [[ "$has" -eq 0 ]] && continue

    echo -e "\n${BOLD}${CYAN}━━━ $c ━━━${NC}"

    for idx in "${!sel_ref[@]}"; do
      [[ "${sel_ref[$idx]}" != "1" ]] && continue
      IFS='|' read -r cat name binary func needs via <<< "${TOOLS[$idx]}"
      [[ "$cat" != "$c" ]] && continue

      if exists "$binary"; then
        echo -e "  ${GREEN}✓${NC} $name — already installed"
        skip_count=$((skip_count + 1))
        continue
      fi

      # Check if runtime dependency is available
      if [[ -n "$needs" ]]; then
        local need_bin
        need_bin=$(need_check "$needs")
        if [[ -n "$need_bin" ]] && ! exists "$need_bin"; then
          echo -e "  ${YELLOW}⚠${NC} $name — requires $needs (not installed). Select Runtimes category first.${NC}"
          fail_count=$((fail_count + 1))
          continue
        fi
      fi

      echo -e "\n  ${BOLD}$name  [${via}]${NC}"
      if $func; then
        if exists "$binary"; then
          echo -e "  ${GREEN}✓${NC} $name installed"
          auto_stow "$binary" "$name"
          ok_count=$((ok_count + 1))
        else
          echo -e "  ${YELLOW}⚠${NC} $name — installed but not in PATH"
          fail_count=$((fail_count + 1))
        fi
      else
        echo -e "  ${RED}✗${NC} $name — failed"
        fail_count=$((fail_count + 1))
      fi
    done
  done

  echo -e "\n${BOLD}${BLUE}━━━ Summary ━━━${NC}"
  echo -e "  ${GREEN}✓${NC} Installed:  $ok_count"
  echo -e "  ${YELLOW}⏭${NC} Skipped:    $skip_count (already existed)"
  echo -e "  ${RED}✗${NC} Failed:     $fail_count"
  echo ""
}

# ─── Interactive selection via fzf (if available) ──────────────────────────
select_fzf() {
  local -n result="$1"
  for idx in "${!TOOLS[@]}"; do result[$idx]="0"; done

  # Step 1: pick category(ies)
  local cats
  cats=$(get_categories)
  local cat_count
  cat_count=$(echo "$cats" | wc -l)

  # Get installed/missing counts per category for display
  local cat_menu=""
  while IFS= read -r c; do
    local total=0 installed=0
    for entry in "${TOOLS[@]}"; do
      IFS='|' read -r ecat ename ebin efunc <<< "$entry"
      if [[ "$ecat" == "$c" ]]; then
        total=$((total + 1))
        exists "$ebin" && installed=$((installed + 1))
      fi
    done
    local missing=$((total - installed))
    if [[ "$missing" -eq 0 ]]; then
      cat_menu+="$c|${GREEN}✓${NC} $c (all $total installed)|$c"$'\n'
    else
      cat_menu+="$c|${YELLOW}${missing}${NC} missing in ${CYAN}$c${NC}|$c"$'\n'
    fi
  done <<< "$cats"

  # Use --delimiter and --with-nth to show only the display column
  local chosen_cats
  chosen_cats=$(
    echo -e "$cat_menu" |
    awk -F'|' '{print $1 "\t" $2}' |
    fzf --multi \
        --prompt="Select categories > " \
        --header="TAB: toggle  |  ENTER: confirm  |  select ALL for all" \
        --delimiter="\t" \
        --with-nth=2 \
        --ansi \
        --layout=reverse-list \
        --bind="ctrl-a:select-all" \
        2>/dev/null || true
  )

  [[ -z "$chosen_cats" ]] && return 0

  # Mark all tools in selected categories
  while IFS= read -r line; do
    local selected_cat
    selected_cat=$(echo "$line" | awk -F'\t' '{print $1}')
    for idx in "${!TOOLS[@]}"; do
      IFS='|' read -r cat name binary func <<< "${TOOLS[$idx]}"
      if [[ "$cat" == "$selected_cat" ]]; then
        result[$idx]="1"
      fi
    done
  done <<< "$chosen_cats"

  # Step 2: fine-tune — show only tools from selected categories
  local any_selected=0
  for idx in "${!result[@]}"; do
    [[ "${result[$idx]}" == "1" ]] && { any_selected=1; break; }
  done
  [[ "$any_selected" -eq 0 ]] && return 0

  # Build tool menu with tab-separated data: INDEX\tDISPLAY
  local tool_menu=""
  local cats_order=()
  while IFS= read -r c; do cats_order+=("$c"); done < <(get_categories)

  local prev_cat=""
  for c in "${cats_order[@]}"; do
    local first=1
    for idx in "${!TOOLS[@]}"; do
      IFS='|' read -r cat name binary func <<< "${TOOLS[$idx]}"
      [[ "$cat" != "$c" ]] && continue
      if [[ "${result[$idx]}" != "1" ]]; then
        local ico; ico=$(status_icon "$binary")
        tool_menu+="$idx\t  ${ico} ${name}\n"
      else
        local ico; ico=$(status_icon "$binary")
        tool_menu+="$idx\t  ${ico} ${name}\n"
      fi
    done
  done

  # Use fzf to fine-tune: selected ones are pre-filled
  local final_selection
  final_selection=$(
    echo -e "$tool_menu" |
    fzf --multi \
        --prompt="Fine-tune selection > " \
        --header="TAB: toggle  |  ENTER: confirm  |  already-installed auto-skipped" \
        --delimiter="\t" \
        --with-nth=2 \
        --ansi \
        --layout=reverse-list \
        2>/dev/null || true
  )

  # Start fresh: only what's in final selection AND is missing
  for idx in "${!result[@]}"; do result[$idx]="0"; done
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local idx
    idx=$(echo "$line" | awk -F'\t' '{print $1}')
    result[$idx]="1"
  done <<< "$final_selection"

  # Auto-unselect already-installed tools (inform user)
  local auto_skipped=0
  for idx in "${!result[@]}"; do
    [[ "${result[$idx]}" != "1" ]] && continue
    IFS='|' read -r cat name binary func <<< "${TOOLS[$idx]}"
    if exists "$binary"; then
      result[$idx]="0"
      auto_skipped=$((auto_skipped + 1))
    fi
  done
  [[ "$auto_skipped" -gt 0 ]] && echo -e "  ${YELLOW}⏭ $auto_skipped already-installed tool(s) auto-deselected${NC}"
}

# ─── Fallback: numbered menu (when fzf is unavailable) ─────────────────────
select_numbered() {
  local -n result="$1"
  for idx in "${!TOOLS[@]}"; do result[$idx]="0"; done

  # Collect categories
  local cats=()
  while IFS= read -r c; do cats+=("$c"); done < <(get_categories)

  # Show overview with status per category
  echo -e "\n${BOLD}Available categories:${NC}"
  local ci=0
  for c in "${cats[@]}"; do
    local total=0 installed=0
    for entry in "${TOOLS[@]}"; do
      IFS='|' read -r ecat ename ebin efunc <<< "$entry"
      [[ "$ecat" == "$c" ]] || continue
      total=$((total + 1))
      exists "$ebin" && installed=$((installed + 1))
    done
    local missing=$((total - installed))
    if [[ "$missing" -eq 0 ]]; then
      echo -e "  ${GREEN}✓${NC} ${CYAN}$c${NC} — all $total installed"
    else
      echo -e "  ${RED}${missing}${NC} missing in ${CYAN}$c${NC} ($installed/$total installed)"
    fi
  done

  # Show detailed per-category tool listing
  echo -e "\n${BOLD}Tools by category:${NC}"
  for c in "${cats[@]}"; do
    echo -e "\n${CYAN}── $c ──${NC}"
    for entry in "${TOOLS[@]}"; do
      IFS='|' read -r ecat ename ebin efunc needs evia <<< "$entry"
      [[ "$ecat" != "$c" ]] && continue
      local icon; icon=$(status_icon "$ebin")
      printf "  %s %-12s %b\n" "$icon" "$ename" "${YELLOW}[$evia]${NC}"
    done
  done

  echo ""
  echo -e "${YELLOW}Pick categories to install (comma-separated), or 'all', or 'q':${NC}"
  echo -ne "${BOLD}>${NC} "
  read -r cat_input

  case "$cat_input" in
    q|quit) echo "Bye."; exit 0 ;;
    all|ALL)
      for idx in "${!TOOLS[@]}"; do result[$idx]="1"; done
      ;;
    *)
      IFS=',' read -ra selected_cats <<< "$cat_input"
      for sc in "${selected_cats[@]}"; do
        sc=$(echo "$sc" | xargs)
        for idx in "${!TOOLS[@]}"; do
          IFS='|' read -r cat name binary func <<< "${TOOLS[$idx]}"
          [[ "$cat" == "$sc" ]] && result[$idx]="1"
        done
      done
      ;;
  esac

  # Fine-tune: toggle individual tools
  local any_selected=0
  for idx in "${!result[@]}"; do
    [[ "${result[$idx]}" == "1" ]] && { any_selected=1; break; }
  done
  [[ "$any_selected" -eq 0 ]] && return 0

  echo -e "\n${YELLOW}Fine-tune selection (optional):${NC}"
  local prev_cat=""
  local tool_list=()
  local j=0
  for idx in "${!TOOLS[@]}"; do
    IFS='|' read -r cat name binary func needs via <<< "${TOOLS[$idx]}"
    [[ "$prev_cat" != "$cat" ]] && echo -e "\n  ${CYAN}$cat:${NC}" && prev_cat="$cat"
    local mark; [[ "${result[$idx]}" == "1" ]] && mark="${GREEN}[x]${NC}" || mark="${RED}[ ]${NC}"
    echo -e "    $j: $mark $name  ${YELLOW}[$via]${NC}"
    tool_list[$j]="$idx"
    j=$((j + 1))
  done
  echo ""
  echo -ne "Toggle numbers (space-separated, e.g. 1 3 5), or press Enter to proceed: "
  read -r toggle_input
  for num in $toggle_input; do
    local ti="${tool_list[$num]}"
    [[ -z "$ti" ]] && continue
    if [[ "${result[$ti]}" == "1" ]]; then result[$ti]="0"; else result[$ti]="1"; fi
  done

  # Auto-unselect installed
  local skipped=0
  for idx in "${!result[@]}"; do
    [[ "${result[$idx]}" != "1" ]] && continue
    IFS='|' read -r cat name binary func <<< "${TOOLS[$idx]}"
    if exists "$binary"; then
      result[$idx]="0"
      skipped=$((skipped + 1))
    fi
  done
  [[ "$skipped" -gt 0 ]] && echo -e "  ${YELLOW}⏭ $skipped already-installed tool(s) removed from selection${NC}"
}

# ─── Interactive entry (numbered menu) ──────────────────────────────────────
interactive() {
  echo -e "${BOLD}${BLUE}━━━ Dotfiles Dependency Installer ━━━${NC}"
  echo -e "Select what you want to install. Already-installed tools are shown with ${GREEN}✓${NC}.\n"

  local -A selection
  select_numbered selection
  show_selection selection
  install_selected selection
}

# ─── Interactive entry (fzf) ────────────────────────────────────────────────
interactive_fzf() {
  echo -e "${BOLD}${BLUE}━━━ Dotfiles Dependency Installer (fzf) ━━━${NC}"
  echo -e "Select what you want to install. Already-installed tools are shown with ${GREEN}✓${NC}.\n"

  local -A selection
  select_fzf selection
  show_selection selection
  install_selected selection
}

# ─── Non-interactive: install all missing ────────────────────────────────────
install_all_missing() {
  echo -e "${BOLD}${BLUE}━━━ Installing all missing dependencies ─━━${NC}\n"

  local -A all_sel
  for idx in "${!TOOLS[@]}"; do
    all_sel[$idx]="1"
  done

  install_selected all_sel
}

# ─── CLI ─────────────────────────────────────────────────────────────────────
case "${1:-}" in
  --list|-l)
    list_all
    ;;
  --install|-i)
    install_all_missing
    ;;
  --fzf)
    if exists fzf; then
      interactive_fzf
    else
      echo -e "${RED}fzf not found. Install it or use the default interactive mode.${NC}" >&2
      exit 1
    fi
    ;;
  --help|-h)
    echo "Usage: $SELF [OPTION]"
    echo ""
    echo "Options:"
    echo "  (no args)    Interactive mode — pick what to install"
    echo "  --fzf        Use fzf picker instead of numbered menu"
    echo "  --list, -l   List all tools with installation status"
    echo "  --install,-i Install all missing tools (non-interactive)"
    echo "  --help, -h   Show this help"
    ;;
  *)
    interactive
    ;;
esac
