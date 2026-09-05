# Cargo/rustup env — only source when rustup actually installed (avoids
# "source: No such file or directory" on rustless WSL boxes).
if test -f "$HOME/.cargo/env.fish"
    source "$HOME/.cargo/env.fish"
end