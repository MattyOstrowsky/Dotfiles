# DevOps Workstation — Dependencies

Flat list of install commands. Run what you need, skip what you have.

## Core CLI Tools

```bash
sudo apt install -y fzf bat jq ripgrep fd-find tree htop direnv
curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh       # zoxide — better cd
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq  # yq — YAML processor
pip3 install --user tldr                                                                     # tldr — simplified man pages
```

## Shell

```bash
# fish
sudo apt-add-repository -y ppa:fish-shell/release-3 && sudo apt update && sudo apt install -y fish

# starship prompt
curl -sS https://starship.rs/install.sh | sh -s -- -y

# fisher (run inside fish)
curl -sL https://raw.githubusercontent.com/jorgebucaran/fisher/main/functions/fisher.fish | source && fisher install jorgebucaran/fisher
fisher update   # installs plugins from fish_plugins
```

## OpenCode

```bash
curl -fsSL https://opencode.ai/install | bash
```

## Neovim + NvChad

```bash
# neovim (latest from source/appimage for v0.12+)
curl -LO https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.appimage
chmod +x nvim-linux-x86_64.appimage
sudo mv nvim-linux-x86_64.appimage /usr/local/bin/nvim

# NvChad bootstraps automatically on first launch from init.lua
# Just stow the config and run nvim — it will install everything
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

# k9s — K8s TUI
curl -sSfL https://github.com/derailed/k9s/releases/latest/download/k9s_Linux_amd64.tar.gz | sudo tar xz -C /usr/local/bin k9s

# kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash && sudo mv kustomize /usr/local/bin/

# stern — multi-pod log tailing
go install github.com/stern/stern@latest

# kubeconform — manifest validation
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

# ansible
sudo apt install -y ansible

# tfsec
go install github.com/aquasecurity/tfsec/cmd/tfsec@latest

# checkov
pip3 install --user checkov

# infracost
curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
```

## Development

```bash
# go — download from https://go.dev/dl/
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz && rm go1.22.2.linux-amd64.tar.gz

# python3
sudo apt install -y python3 python3-pip python3-venv

# bun
curl -fsSL https://bun.sh/install | bash

# git
sudo apt install -y git
```

## TUI Tools

```bash
# lazygit
go install github.com/jesseduffield/lazygit@latest

# lazydocker
go install github.com/jesseduffield/lazydocker@latest

# lazysql — database TUI (Postgres, MySQL, SQLite)
go install github.com/jorgerojas26/lazysql@latest

# dive — Docker image layer explorer
DIVEVER=$(curl -s https://api.github.com/repos/wagoodman/dive/releases/latest | jq -r .tag_name)
wget -qO /tmp/dive.deb "https://github.com/wagoodman/dive/releases/download/${DIVEVER}/dive_${DIVEVER#v}_linux_amd64.deb"
sudo dpkg -i /tmp/dive.deb && rm /tmp/dive.deb

# btop — system monitor
sudo apt install -y btop

# ctop — container metrics
sudo wget -qO /usr/local/bin/ctop https://github.com/bcicen/ctop/releases/latest/download/ctop-linux-amd64
sudo chmod +x /usr/local/bin/ctop

# glow — markdown renderer
go install github.com/charmbracelet/glow@latest

# navi — interactive cheatsheet (Ctrl+G)
cargo install --locked navi

# atuin — better shell history with sync
curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | sh
```

## Security Scanners

```bash
# trivy
sudo apt-get install -y wget apt-transport-https gnupg
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt-get update && sudo apt-get install -y trivy

# grype
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
```

## Post-Install

```bash
# apply dotfiles
cd ~/Dotfiles
stow fish starship opencode nvim

# install fish plugins (run inside fish)
fisher update

# set fish as default shell
chsh -s $(which fish)
```
