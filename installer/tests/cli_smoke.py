#!/usr/bin/env python3
"""CLI contract: JSON, strict input, no writes during planning, idempotence."""
import json
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parents[2]
binary = repo / "installer/dotfiles-install"

def run(*args, ok=True):
    result = subprocess.run([str(binary), *map(str, args), "--json"], cwd=repo, text=True, capture_output=True)
    data = json.loads(result.stdout)
    assert (result.returncode == 0) == ok, (result.returncode, data, result.stderr)
    if not ok:
        assert data.get("error"), data
    return data

with tempfile.TemporaryDirectory(prefix="dotfiles-cli-") as temp:
    target = Path(temp) / "home"
    args = ("--profile", "installer/examples/config-only.json", "--target", target)
    assert run(*args, "--dry-run")["changed"]
    assert not target.exists(), "dry-run created target"
    assert run(*args, "--yes")["changed"]
    assert not run(*args, "--yes")["changed"]
    assert not run(*args, "--dry-run")["changed"]
    assert (target / ".config/fish/config.fish").is_file()
    run("--yes", ok=False)
    run("--tools", "does-not-exist", "--yes", ok=False)
    run("--tools", "git", "--packages=--evil", "--dry-run", ok=False)
    run("--tools", "git", "--no-configs", "--configs-only", "--yes", ok=False)
    run("--profile", "installer/examples/config-only.json", ok=False)
    profile = Path(temp) / "typo.yaml"
    profile.write_text("version: 1\nconfigz: false\n")
    run("--profile", profile, "--yes", ok=False)
    # CLI selection overrides replace profile.tools instead of adding to it.
    other = Path(temp) / "other"
    run("--profile", "installer/examples/config-only.json", "--tools", "git", "--target", other, "--yes")
    assert (other / ".config/git/config").exists()
    assert not (other / ".config/fish").exists()
print("PASS CLI: strict profiles, JSON failures, overrides, dry-run and idempotence")
