package main

// This file is the auxiliary, planner-facing model for the config/ stow
// tree. It deliberately stays out of the manifest (tools.yaml): the manifest
// only describes installable tools, while host scope and package intent for
// the shared config tree live here so Ubuntu/Debian, Fedora, Arch and WSL
// installers can filter and dry-run before touching anything.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HostScope says which hosts a config package applies to.
type HostScope string

const (
	// HostAll is the default: the package is distro/host agnostic and safe
	// on Ubuntu/Debian, Fedora, Arch and WSL alike.
	HostAll HostScope = "all"
	// HostWSL marks Windows-side packages of a WSL setup. Their content
	// belongs to the Windows host and must not be stowed into the Linux
	// home directory.
	HostWSL HostScope = "wsl"
)

// ConfigPkg is the auxiliary metadata of one config package. The registry
// only pins what the plain tree scan cannot know (host scope, intent);
// everything else defaults to HostAll.
type ConfigPkg struct {
	Name string
	Host HostScope // scope marker (HostAll unless overridden in the registry)
	Desc string    // short planner/UI description
}

// configPkgMeta is the authoritative auxiliary registry, keyed by directory
// name in config/. Directories without an entry still scan as packages with
// HostAll defaults; add an entry here when a package needs scope or docs.
var configPkgMeta = map[string]ConfigPkg{
	"atuin":            {Desc: "atuin shell-history config"},
	"bat":              {Desc: "bat pager theme and config"},
	"btop":             {Desc: "btop system-monitor config"},
	"direnv":           {Desc: "direnv hook config"},
	"dive":             {Desc: "dive image-analyzer config"},
	"fish":             {Desc: "fish shell config (prompt, plugins, completions)"},
	"glow":             {Desc: "glow markdown renderer config"},
	"k9s":              {Desc: "k9s kubernetes client config and skins"},
	"lazydocker":       {Desc: "lazydocker TUI config"},
	"lazygit":          {Desc: "lazygit TUI config"},
	"navi":             {Desc: "navi cheatsheets"},
	"nvim":             {Desc: "nvim (NvChad) config"},
	"omp":              {Desc: "oh-my-posh agent config"},
	"opencode":         {Desc: "opencode editor config"},
	"starship":         {Desc: "starship prompt config"},
	"windows-terminal": {Host: HostWSL, Desc: "Windows Terminal color schemes and profile notes (Windows side of WSL)"},
}

// pkgHostScope returns the host scope for a package directory name.
func pkgHostScope(name string) HostScope {
	meta, ok := configPkgMeta[name]
	if !ok || meta.Host == "" {
		return HostAll
	}
	return meta.Host
}

// metaForPkg returns the registry entry for name, defaulting to HostAll for
// tree additions that have no entry yet.
func metaForPkg(name string) ConfigPkg {
	meta, ok := configPkgMeta[name]
	if !ok {
		return ConfigPkg{Name: name, Host: HostAll}
	}
	if meta.Host == "" {
		meta.Host = HostAll
	}
	meta.Name = name
	return meta
}

// ScanConfigPkgs lists config packages in <repoRoot>/config/ together with
// their auxiliary metadata. It shares the tree scan with scanStowPkgs:
// stray files at the tree root, dot-entries and empty directories are
// explicitly not packages.
func ScanConfigPkgs(repoRoot string) ([]ConfigPkg, error) {
	names, err := stowTreePackageDirs(configDir(repoRoot))
	if err != nil {
		return nil, err
	}
	pkgs := make([]ConfigPkg, 0, len(names))
	for _, name := range names {
		pkgs = append(pkgs, metaForPkg(name))
	}
	return pkgs, nil
}

// PlanAction is a per-file outcome computed by PlanStow without touching
// the filesystem.
type PlanAction string

const (
	// PlanLink: the target is free, stow will create the link.
	PlanLink PlanAction = "link"
	// PlanRestow: the target is already owned by the repo stow tree,
	// stow refreshes/repairs the link.
	PlanRestow PlanAction = "restow"
	// PlanBackup: the target is occupied by non-repo content, it would be
	// moved to the backup dir before stowing.
	PlanBackup PlanAction = "backup"
)

// PlanFile describes the planned outcome for one package-relative file.
type PlanFile struct {
	Rel     string     // package-relative path inside config/<pkg>
	Target  string     // absolute target path stow would manage
	Action  PlanAction // what applyStow would do
	Managed bool       // true when the repo currently owns the target (PlanRestow)
	Detail  string     // human note for backup actions (what occupies the target)
}

// StowPlan is the dry run of applyStow for one package: nothing on disk is
// touched. Conflicts are only reported so preflight can warn before any
// backup happens.
type StowPlan struct {
	Pkg        string
	Target     string
	Files      []PlanFile
	BackupRels []string // package-relative paths that would be backed up
	BackupDir  string   // ~/.dotfiles-backup/<stamp>/ when BackupRels is non-empty
}

// HasConflicts reports whether the plan would back anything up.
func (s *StowPlan) HasConflicts() bool { return len(s.BackupRels) > 0 }

// PlanStow computes what applyStow would do for p against target without
// mutating the filesystem. target may use a leading ~. The result matches
// applyStow's behaviour because both share StowPkg.classifyTarget.
func PlanStow(p *StowPkg, target string) *StowPlan {
	target = expandTarget(target)
	plan := &StowPlan{Pkg: p.Name, Target: target}
	for _, f := range p.Files {
		tp := filepath.Join(target, f)
		pf := PlanFile{Rel: f, Target: tp}
		switch p.classifyTarget(tp) {
		case targetAbsent:
			pf.Action = PlanLink
		case targetManaged:
			pf.Action = PlanRestow
			pf.Managed = true
		default:
			pf.Action = PlanBackup
			pf.Detail = describeTarget(tp)
			plan.BackupRels = append(plan.BackupRels, f)
		}
		plan.Files = append(plan.Files, pf)
	}
	if plan.HasConflicts() {
		home, _ := os.UserHomeDir()
		plan.BackupDir = filepath.Join(home, ".dotfiles-backup", time.Now().Format("20060102-150405"))
	}
	return plan
}

// describeTarget summarizes what currently occupies tp, for preflight
// messages about planned backups.
func describeTarget(tp string) string {
	fi, err := os.Lstat(tp)
	if err != nil {
		return "missing"
	}
	switch {
	case fi.Mode()&os.ModeSymlink != 0:
		dest, rerr := os.Readlink(tp)
		if rerr != nil {
			return "unreadable symlink"
		}
		if _, serr := os.Stat(tp); serr != nil {
			return fmt.Sprintf("dangling symlink → %s", dest)
		}
		return fmt.Sprintf("symlink → %s", dest)
	case fi.IsDir():
		return "directory"
	default:
		return "file"
	}
}
