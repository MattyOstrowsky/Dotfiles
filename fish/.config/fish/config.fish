# =============================================================================
# Fish Shell Config — DevOps Workstation
# =============================================================================

# Editor
set -gx EDITOR nvim

# PATH
set -gx PATH $HOME/.local/bin $PATH

# Go
set -gx GOPATH $HOME/go
set -gx PATH $GOPATH/bin /usr/local/go/bin $PATH

# Cargo (Rust tools: navi, etc.)
set -gx PATH $HOME/.cargo/bin $PATH

# fnm (Fast Node Manager)
set -gx PATH $HOME/.local/share/fnm $PATH

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

    # atuin (better shell history with sync)
    if type -q atuin
        atuin init fish | source
    end

    # navi (interactive cheatsheet — Ctrl+G)
    if type -q navi
        navi widget fish | source
    end

    # fnm (Node.js version manager)
    if type -q fnm
        fnm env --use-on-cd | source
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
alias kgpa='kubectl get pods -A'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kgn='kubectl get nodes'
alias kgi='kubectl get ingress'
alias kgcm='kubectl get configmaps'
alias kgsec='kubectl get secrets'
alias klog='kubectl logs -f'
alias kex='kubectl exec -it'
alias kdesc='kubectl describe'
alias kaf='kubectl apply -f'
alias kdf='kubectl delete -f'
alias kroll='kubectl rollout restart deployment'
alias kevents='kubectl get events --sort-by=.lastTimestamp'

# kubectx / kubens — fast context & namespace switching
alias kctx='kubectx'
alias kns='kubens'

# k9s — terminal K8s dashboard
alias k9='k9s'

# stern — multi-pod log tailing
alias ksl='stern'

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
alias dcu='docker compose up -d'
alias dcd='docker compose down'
alias dcl='docker compose logs -f'
alias dps='docker ps'
alias dpsa='docker ps -a'
alias dimg='docker images'
alias dlog='docker logs -f'
alias dex='docker exec -it'
alias dprune='docker system prune -af'

# lazydocker / lazygit / lazysql — terminal UIs
alias lzd='lazydocker'
alias lzg='lazygit'
alias lzs='lazysql'

# dive — explore docker image layers
alias ddive='dive'

# btop — system monitor
alias top='btop'

# ctop — container monitor
alias dtop='ctop'

# glow — render markdown in terminal
alias md='glow'

# =============================================================================
# Aliases — Security / Scanning
# =============================================================================
alias tscan='trivy image'
alias tfs-scan='tfsec .'

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

# Watch pods in namespace
function kwatch
    kubectl get pods -w $argv
end

# Get all resources in namespace
function kall
    kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found $argv
end

# Decode K8s secret
function ksecret
    kubectl get secret $argv[1] -o jsonpath='{.data}' | jq 'to_entries[] | {key: .key, value: (.value | @base64d)}'
end

# Quick terraform workspace switch
function tfw-switch
    terraform workspace select $argv[1] || terraform workspace new $argv[1]
end

# Docker compose rebuild specific service
function dcrebuild
    docker compose up -d --build $argv[1]
end

# Show listening ports
function ports
    sudo ss -tlnp | grep LISTEN
end

# =============================================================================
# Dotfiles auto-update — runs once per shell session
# =============================================================================
function __dotfiles_check_update --on-event fish_prompt
    # One-shot: skip on subsequent prompts
    set -q __dotfiles_update_done; and return
    set -g __dotfiles_update_done 1

    # Guard: Dotfiles repo must exist
    if not test -d "$HOME/Dotfiles/.git"
        return
    end

    # Guard: git and make must be installed
    if not command -q git; or not command -q make
        return
    end

    # Fetch silently, abort on failure (offline, no remote, etc.)
    if not git -C "$HOME/Dotfiles" fetch --quiet 2>/dev/null
        return
    end

    # Count commits behind upstream (fallback to 0 on error)
    set -l behind (git -C "$HOME/Dotfiles" rev-list --count HEAD..@{upstream} 2>/dev/null)
    or set behind 0

    if test "$behind" -gt 0
        echo "📦 Dotfiles: $behind update(s) available"
        if git -C "$HOME/Dotfiles" pull --ff-only --quiet 2>/dev/null
            and make -C "$HOME/Dotfiles" install --quiet 2>/dev/null
            echo "✅ Dotfiles updated"
        else
            echo "⚠️  Dotfiles update failed — check manually"
        end
    end
end
