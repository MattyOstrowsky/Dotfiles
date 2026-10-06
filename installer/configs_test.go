package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) ([]Tool, Profile, Options) {
	t.Helper()
	repo := t.TempDir()
	target := t.TempDir()
	src := filepath.Join(repo, "config", "fish", ".config", "fish", "config.fish")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := defaultProfile()
	p.Install = false
	return []Tool{{Name: "fish", Config: "fish"}}, p, Options{Repo: repo, Target: target, BinDir: filepath.Join(target, ".local/bin")}
}
func TestCopyBackupAndIdempotence(t *testing.T) {
	tools, p, o := fixture(t)
	dest := filepath.Join(o.Target, ".config/fish/config.fish")
	os.MkdirAll(filepath.Dir(dest), 0o755)
	os.WriteFile(dest, []byte("original"), 0o600)
	r := Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if !r.Changed {
		t.Fatal("expected change")
	}
	backups, _ := filepath.Glob(filepath.Join(filepath.Dir(dest), ".dotfiles-backup-*"))
	if len(backups) != 1 {
		t.Fatal(backups)
	}
	old, _ := os.ReadFile(backups[0])
	if string(old) != "original" {
		t.Fatal("backup content")
	}
	r = Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if r.Changed {
		t.Fatal("second apply changed")
	}
}
func TestDryRunNoWritesAndConflict(t *testing.T) {
	tools, p, o := fixture(t)
	o.Target = filepath.Join(o.Target, "nonexistent")
	o.DryRun = true
	r := Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(o.Target); !os.IsNotExist(err) {
		t.Fatal("dry run wrote target")
	}
	if !r.Changed {
		t.Fatal("expected predicted change")
	}
	dest := filepath.Join(o.Target, ".config/fish/config.fish")
	os.MkdirAll(filepath.Dir(dest), 0o755)
	os.WriteFile(dest, []byte("mine"), 0o644)
	p.Conflict = "error"
	if _, err := planConfigs(tools, p, o); err == nil {
		t.Fatal("expected conflict")
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "mine" {
		t.Fatal("changed conflict")
	}
}
func TestLinksAndStowMigration(t *testing.T) {
	tools, p, o := fixture(t)
	sourceDir := filepath.Join(o.Repo, "config/fish/.config/fish")
	destDir := filepath.Join(o.Target, ".config/fish")
	os.MkdirAll(filepath.Dir(destDir), 0o755)
	os.Symlink(sourceDir, destDir)
	os.WriteFile(filepath.Join(sourceDir, "fish_variables"), []byte("local state"), 0o600)
	r := Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if !r.Changed {
		t.Fatal("expected detach")
	}
	info, _ := os.Lstat(destDir)
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("still stowed")
	}
	state, err := os.ReadFile(filepath.Join(destDir, "fish_variables"))
	if err != nil || string(state) != "local state" {
		t.Fatal("lost runtime state")
	}
	os.WriteFile(filepath.Join(destDir, "config.fish"), []byte("local"), 0o644)
	data, _ := os.ReadFile(filepath.Join(sourceDir, "config.fish"))
	if string(data) != "new" {
		t.Fatal("wrote into repository")
	}
	p.ConfigMode = "link"
	r = Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	r = Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if r.Changed {
		t.Fatal("link apply not idempotent")
	}
}
func TestUnrelatedSymlinkParentRejected(t *testing.T) {
	tools, p, o := fixture(t)
	other := t.TempDir()
	os.Symlink(other, filepath.Join(o.Target, ".config"))
	if _, err := planConfigs(tools, p, o); err == nil {
		t.Fatal("followed unrelated parent link")
	}
	files, _ := os.ReadDir(other)
	if len(files) > 0 {
		t.Fatal("wrote through symlink")
	}
}
func TestNoConfigsIgnoresMissingRepo(t *testing.T) {
	tools, p, o := fixture(t)
	p.Configs = false
	o.Repo = "/does-not-exist"
	plan, err := planConfigs(tools, p, o)
	if err != nil || len(plan.Files) > 0 {
		t.Fatal(plan, err)
	}
}
func TestFileSymlinkBackedUpWithoutModifyingItsTarget(t *testing.T) {
	tools, p, o := fixture(t)
	outside := filepath.Join(t.TempDir(), "secret")
	os.WriteFile(outside, []byte("private"), 0o600)
	dest := filepath.Join(o.Target, ".config/fish/config.fish")
	os.MkdirAll(filepath.Dir(dest), 0o755)
	os.Symlink(outside, dest)
	r := Result{}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(outside)
	if string(content) != "private" {
		t.Fatal("modified link target")
	}
}
