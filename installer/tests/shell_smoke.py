#!/usr/bin/env python3
"""Exercise actual fish/Starship with isolated HOME, without installing packages."""
import os
import re
from pathlib import Path
import shutil
import socket
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
fish = shutil.which("fish")
starship = shutil.which("starship")
if not fish:
    print("SKIP shell smoke: fish is not installed")
    raise SystemExit(0)
with tempfile.TemporaryDirectory(prefix="dotfiles-shell-") as temp:
    home = Path(temp)
    shutil.copytree(repo / "config/fish/.config/fish", home / ".config/fish", ignore=shutil.ignore_patterns("fish_variables"))
    env = dict(os.environ, HOME=temp, XDG_CONFIG_HOME=str(home / ".config"), XDG_DATA_HOME=str(home / ".local/share"), XDG_CACHE_HOME=str(home / ".cache"), PATH="/usr/bin:/bin:/usr/bin", TERM="xterm-256color")
    for key in ("GOPATH", "CARGO_HOME", "BUN_INSTALL", "FISH_HISTORY", "FNM_DIR", "FNM_MULTISHELL_PATH"):
        env.pop(key, None)
    def run(command, interactive=False):
        return subprocess.run([fish, "-ic" if interactive else "-c", command], env=env, text=True, capture_output=True, check=True)
    first = run('printf "%s\\n" $PATH').stdout.splitlines()
    nested = run("fish -c 'printf \"%s\\n\" $PATH'").stdout.splitlines()
    assert first == nested, (first, nested)
    assert len(first) == len(set(first)), first
    assert first[:2] == ["/usr/bin", "/bin"], first
    assert not any(path.startswith(temp) for path in first), first
    result = run("for cmd in cat bat grep top; if functions -q $cmd; functions --details $cmd; end; end", True)
    assert all(line.startswith(("/usr/share/fish/", "embedded:functions/")) for line in result.stdout.splitlines()), result.stdout
    assert not result.stderr.strip(), result.stderr
    assert not (home / ".config/fish/fish_variables").exists() or "fish_user_paths" not in (home / ".config/fish/fish_variables").read_text()
    env["EDITOR"] = "custom-editor"
    assert run('echo $EDITOR').stdout.strip() == "custom-editor"
    print("PASS fish: nested PATH stable, missing tools tolerated, commands unshadowed, EDITOR preserved")
    if starship:
        env["STARSHIP_CONFIG"] = str(repo / "config/starship/.config/starship.toml")
        result = subprocess.run([starship, "prompt", "--path", temp], env=env, text=True, capture_output=True, check=True)
        assert socket.gethostname() not in result.stdout, result.stdout
        assert all(part in result.stdout for part in ("┌─>", "└─>")), result.stdout
        plain_prompt = re.sub(r"\x1b\[[0-?]*[ -/]*[@-~]", "", result.stdout)
        assert len([line for line in plain_prompt.splitlines() if line.strip()]) == 2, result.stdout
        assert not result.stderr.strip(), result.stderr
        local_bin = home / ".local/bin"
        local_bin.mkdir(parents=True)
        (local_bin / "starship").symlink_to(starship)
        definition = run("functions fish_prompt", True)
        assert "starship prompt" in definition.stdout, definition.stdout
        result = run("fish_prompt", True)
        assert socket.gethostname() not in result.stdout, result.stdout
        assert all(part in result.stdout for part in ("┌─>", "└─>")), result.stdout
        assert not result.stderr.strip(), result.stderr
        print("PASS Starship: fish initializes local binary, prompt renders two lines without hostname")
    else:
        print("SKIP Starship: binary missing")
