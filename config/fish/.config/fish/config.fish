# Shared fish config: no downloads, package installs or universal PATH writes.
# Keep inherited environment paths (venvs, SSH, WSL); remove exact duplicates.
set -l clean_path
for entry in $PATH
    if test -n "$entry"; and not contains -- "$entry" $clean_path
        set -a clean_path "$entry"
    end
end
set -gx PATH $clean_path

# Add optional tool directories once, in a predictable order. Respect custom
# GOPATH/CARGO_HOME/BUN_INSTALL; do not export machine-specific defaults.
set -l go_home "$HOME/go"
set -q GOPATH; and set go_home (string split : -- "$GOPATH")
set -l cargo_home "$HOME/.cargo"
set -q CARGO_HOME; and set cargo_home "$CARGO_HOME"
set -l bun_home "$HOME/.bun"
set -q BUN_INSTALL; and set bun_home "$BUN_INSTALL"
set -l optional_paths "$HOME/.local/bin" "$cargo_home/bin" "$bun_home/bin" "$HOME/.local/share/fnm" "$HOME/.atuin/bin" /usr/local/go/bin
for dir in $go_home
    set -a optional_paths "$dir/bin"
end
# --path avoids persistent fish_user_paths; --append preserves activated venvs.
for dir in $optional_paths
    if test -d "$dir"
        fish_add_path --global --path --append "$dir"
    end
end

if not set -q EDITOR
    if command -q nvim
        set -gx EDITOR nvim
    else if command -q nano
        set -gx EDITOR nano
    else
        set -gx EDITOR vi
    end
end
set -q VISUAL; or set -gx VISUAL "$EDITOR"

if status is-interactive
    set -g fish_greeting

    # Nord syntax and completion colors, without machine-specific fish_variables.
    set -g fish_color_normal d8dee9
    set -g fish_color_command 88c0d0
    set -g fish_color_param e5e9f0
    set -g fish_color_keyword 81a1c1
    set -g fish_color_quote a3be8c
    set -g fish_color_redirection b48ead
    set -g fish_color_end 81a1c1
    set -g fish_color_error bf616a
    set -g fish_color_comment 616e88
    set -g fish_color_autosuggestion 616e88
    set -g fish_color_operator 81a1c1
    set -g fish_color_escape ebcb8b
    set -g fish_color_search_match --background=434c5e
    set -g fish_color_selection --background=434c5e
    set -g fish_pager_color_prefix 88c0d0
    set -g fish_pager_color_completion d8dee9
    set -g fish_pager_color_description 81a1c1
    set -g fish_pager_color_selected_background --background=434c5e

    if command -q fzf
        set -gx FZF_DEFAULT_OPTS '--color=bg:#2e3440,fg:#d8dee9,hl:#81a1c1,bg+:#434c5e,fg+:#eceff4,hl+:#88c0d0,info:#ebcb8b,prompt:#88c0d0,pointer:#b48ead,marker:#a3be8c,spinner:#b48ead,header:#81a1c1'
        # Native bindings when provided by the distribution's fish integration.
        if functions -q fzf_key_bindings
            fzf_key_bindings
        end
    end
    if command -q starship
        starship init fish | source
    end
    if command -q zoxide
        zoxide init fish | source
    end
    if command -q direnv
        direnv hook fish | source
    end
    if command -q atuin
        atuin init fish | source
    end
    if command -q fnm
        fnm env --use-on-cd --shell fish | source
    end

    # Explicit shortcuts only: cat, bat, grep, top, cd and ls stay separate.
    abbr -a ll 'ls -lah'
    abbr -a gs 'git status'
    abbr -a ga 'git add'
    abbr -a gc 'git commit'
    abbr -a gd 'git diff'
    abbr -a gl 'git log --oneline -20'
    if command -q docker
        abbr -a d 'docker'
        abbr -a dc 'docker compose'
        abbr -a dps 'docker ps'
    end
    if command -q lazygit
        abbr -a lzg 'lazygit'
    end
    if command -q lazydocker
        abbr -a lzd 'lazydocker'
    end
end
