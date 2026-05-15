# =============================================================================
# Fish Shell Config — DevOps Workstation
# =============================================================================

# Editor
set -gx EDITOR nvim

# PATH
set -gx PATH $HOME/.local/bin $PATH

# Bun
set -gx BUN_INSTALL "$HOME/.bun"
set -gx PATH $BUN_INSTALL/bin $PATH

# =============================================================================
# Interactive Shell
# =============================================================================

if status is-interactive

    # Starship prompt
    starship init fish | source

    # -------------------------------------------------------------------------
    # Tool completions (lazy-loaded)
    # -------------------------------------------------------------------------
    if type -q kubectl
        kubectl completion fish | source
    end

    if type -q helm
        helm completion fish | source
    end

    if type -q terraform
        complete -c terraform -f -a "(terraform -install-autocomplete 2>/dev/null)"
    end

    if type -q docker
        # Docker completions are usually installed with docker
    end

    if type -q aws
        complete -C aws_completer aws
    end

    # direnv (auto-load .envrc)
    if type -q direnv
        direnv hook fish | source
    end

    # zoxide (better cd)
    if type -q zoxide
        zoxide init fish | source
    end

end

# =============================================================================
# Aliases — General
# =============================================================================
alias bat='batcat'
alias n='nvim'
alias cat='batcat --paging=never'
alias ll='ls -lah'
alias ..='cd ..'
alias ...='cd ../..'
alias grep='grep --color=auto'

# =============================================================================
# Aliases — Git
# =============================================================================
alias g='git'
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline -20'
alias gd='git diff'
alias gco='git checkout'
alias gb='git branch'
alias gpl='git pull'

# =============================================================================
# Aliases — Kubernetes
# =============================================================================
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kgn='kubectl get nodes'
alias kns='kubectl config set-context --current --namespace'
alias kctx='kubectl config get-contexts'
alias klog='kubectl logs -f'
alias kex='kubectl exec -it'
alias kdesc='kubectl describe'
alias kaf='kubectl apply -f'
alias kdf='kubectl delete -f'

# =============================================================================
# Aliases — Helm
# =============================================================================
alias h='helm'
alias hl='helm list'
alias hi='helm install'
alias hu='helm upgrade'
alias hd='helm delete'

# =============================================================================
# Aliases — Terraform
# =============================================================================
alias tf='terraform'
alias tfi='terraform init'
alias tfp='terraform plan'
alias tfa='terraform apply'
alias tfv='terraform validate'
alias tff='terraform fmt -recursive'
alias tfs='terraform state list'
alias tfw='terraform workspace'

# =============================================================================
# Aliases — Docker
# =============================================================================
alias d='docker'
alias dc='docker compose'
alias dps='docker ps'
alias dpsa='docker ps -a'
alias dimg='docker images'
alias dlog='docker logs -f'
alias dex='docker exec -it'
alias dprune='docker system prune -af'

# =============================================================================
# Functions — DevOps helpers
# =============================================================================

# Quick switch k8s namespace
function kn
    kubectl config set-context --current --namespace=$argv[1]
    echo "Switched to namespace: $argv[1]"
end

# Port forward shortcut
function kpf
    kubectl port-forward $argv[1] $argv[2]
end

# Docker shell into container
function dsh
    docker exec -it $argv[1] /bin/sh
end

# Terraform plan + apply workflow
function tfpa
    terraform plan -out=plan.tfplan && echo "Plan saved. Run 'terraform apply plan.tfplan' to apply."
end

# Show pod resource usage
function ktop
    kubectl top pods $argv
end

# Quick git commit + push
function gcp
    git add -A && git commit -m "$argv[1]" && git push
end

# Create directory and cd into it
function mkcd
    mkdir -p $argv[1] && cd $argv[1]
end
