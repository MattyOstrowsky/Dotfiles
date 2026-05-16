# Dependencies

Everything needed to bootstrap this DevOps workstation dotfiles repo (GNU Stow).

## One-Shot Install

```bash
sudo apt install -y stow git curl wget make fzf bat ripgrep fd-find tree htop direnv btop python3 python3-pip python3-venv ansible
```

## Per-Package Dependencies

| Package | Required | Optional | Install |
|---------|----------|----------|---------|
| fish | `fish` + `starship` | `fzf`, `bat`, `zoxide`, `direnv`, `atuin`, `navi` | apt + curl + cargo |
| nvim | `nvim` (appimage) + `git` | `ripgrep`, `fd-find`, `grep`, `sed` | appimage |
| opencode | `opencode` | — | curl pipe bash |
| navi | `cargo` (Rust) | — | `cargo install` |
| starship | `starship` | — | curl pipe sh |
| pi | `pi` binary | — | binary download |

### fish — Plugins (9, via `fish_plugins`)

jorgebucaran/fisher, PatrickF1/fzf.fish, jethrokuan/z, jorgebucaran/autopair.fish, meaningful-ooo/sponge, gazorby/fish-abbreviation-tips, nickeb96/puffer-fish, jorgebucaran/replay.fish, laughedelic/pisces

### fish — Aliases Reference

kubectl, helm, terraform, docker, git, aws, k9s, stern, kubectx, lazygit, lazydocker, lazysql, dive, btop, ctop, glow, trivy, tfsec

### nvim — Notes

NvChad bootstraps itself on first launch. Just stow and run `nvim`.

## Shell

```bash
# fish (PPA for latest)
sudo apt-add-repository -y ppa:fish-shell/release-3 && sudo apt update && sudo apt install -y fish

# starship prompt
curl -sS https://starship.rs/install.sh | sh -s -- -y

# zoxide — smarter cd
curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh

# atuin — shell history with sync
curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | sh

# fisher plugins (run after stow fish)
curl -sL https://raw.githubusercontent.com/jorgebucaran/fisher/main/functions/fisher.fish | source && fisher install jorgebucaran/fisher
fisher update
```

## Neovim

```bash
curl -LO https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.appimage
chmod +x nvim-linux-x86_64.appimage && sudo mv nvim-linux-x86_64.appimage /usr/local/bin/nvim
```

## OpenCode

```bash
curl -fsSL https://opencode.ai/install | bash
```

## Language Runtimes

```bash
# Go
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz && rm go1.22.2.linux-amd64.tar.gz

# Rust (cargo)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

# Bun
curl -fsSL https://bun.sh/install | bash

# Python (pip packages)
pip3 install --user checkov tldr

# Navi (interactive cheatsheet)
cargo install --locked navi
```

## Kubernetes

```bash
# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && rm kubectl

# helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# kubectx + kubens
sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens
mkdir -p ~/.config/fish/completions
ln -sf /opt/kubectx/completion/kubectx.fish ~/.config/fish/completions/
ln -sf /opt/kubectx/completion/kubens.fish ~/.config/fish/completions/

# k9s
curl -sSfL https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz | sudo tar xz -C /usr/local/bin k9s

# stern
go install github.com/stern/stern@latest

# kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash && sudo mv kustomize /usr/local/bin/

# kubeconform
go install github.com/yannh/kubeconform/cmd/kubeconform@latest
```

## Infrastructure as Code

```bash
# terraform
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install -y terraform

# terragrunt
TGVER=$(curl -s https://api.github.com/repos/gruntwork-io/terragrunt/releases/latest | jq -r .tag_name)
sudo wget -qO /usr/local/bin/terragrunt "https://github.com/gruntwork-io/terragrunt/releases/download/${TGVER}/terragrunt_linux_amd64"
sudo chmod +x /usr/local/bin/terragrunt

# tfsec
go install github.com/aquasecurity/tfsec/cmd/tfsec@latest

# infracost
curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
```

## TUI Tools

```bash
# lazygit
go install github.com/jesseduffield/lazygit@latest

# lazydocker
go install github.com/jesseduffield/lazydocker@latest

# lazysql
go install github.com/jorgerojas26/lazysql@latest

# dive
DIVEVER=$(curl -s https://api.github.com/repos/wagoodman/dive/releases/latest | jq -r .tag_name)
wget -qO /tmp/dive.deb "https://github.com/wagoodman/dive/releases/download/${DIVEVER}/dive_${DIVEVER#v}_linux_amd64.deb"
sudo dpkg -i /tmp/dive.deb && rm /tmp/dive.deb

# ctop
sudo wget -qO /usr/local/bin/ctop https://github.com/bcicen/ctop/releases/latest/download/ctop-linux-amd64
sudo chmod +x /usr/local/bin/ctop

# glow
go install github.com/charmbracelet/glow@latest
```

## Security

```bash
# trivy
sudo apt-get install -y wget apt-transport-https gnupg
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt-get update && sudo apt-get install -y trivy

# grype
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
```

## Pi CLI

```bash
# Download latest pi binary and place in PATH
# https://github.com/example/pi-releases (adjust URL to actual release)
```

## Post-Install

```bash
# Stow all packages
cd ~/Dotfiles && make install

# Set fish as default shell
chsh -s $(which fish)

# Install fish plugins (run inside fish)
fisher update
```
