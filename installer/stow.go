package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// stowTreeName is the repo directory holding the shared, distro-agnostic
// stow packages (renamed from wsl/ so the tree reads as plain configs).
const stowTreeName = "config"

// configDir returns the repo's shared stow tree (<repo>/config).
func configDir(repoRoot string) string { return filepath.Join(repoRoot, stowTreeName) }

type StowState int

const (
	StowNone StowState = iota
	StowPartial
	StowFull
)

func (s StowState) String() string {
	switch s {
	case StowFull:
		return "stowed"
	case StowPartial:
		return "partial"
	}
	return "not stowed"
}

type StowPkg struct {
	Name     string
	Dir      string    // absolute path to <repo>/config/<name>
	Host     HostScope // where the package applies (HostAll unless host-specific, e.g. windows-terminal)
	Mappings []string  // display mappings, e.g. ~/.config/fish
	Files    []string  // package-relative file paths (stow link targets)
}

// findRepoRoot walks up from the executable (then cwd) until a directory
// containing config/ (the shared stow tree) is found.
func findRepoRoot() (string, error) {
	var starts []string
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	for _, start := range starts {
		dir := start
		for {
			if st, err := os.Stat(filepath.Join(dir, stowTreeName)); err == nil && st.IsDir() {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", fmt.Errorf("could not locate repo root (no %s/ directory found above executable or cwd)", stowTreeName)
}

// scanStowPkgs lists config packages in <repoRoot>/config/.
func scanStowPkgs(repoRoot string) ([]StowPkg, error) {
	names, err := stowTreePackageDirs(configDir(repoRoot))
	if err != nil {
		return nil, err
	}
	pkgs := make([]StowPkg, 0, len(names))
	for _, name := range names {
		dir := filepath.Join(configDir(repoRoot), name)
		files, err := pkgFiles(dir)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, StowPkg{
			Name:     name,
			Dir:      dir,
			Host:     pkgHostScope(name),
			Mappings: pkgMappings(dir),
			Files:    files,
		})
	}
	return pkgs, nil
}

// stowTreePackageDirs returns the names of real config packages directly
// under the stow tree: directories that contain at least one stow-able
// entry. Dot-entries (runtime state, editor swap dirs), stray files at the
// tree root (READMEs, notes, scripts) and empty directories are explicitly
// not config packages. ReadDir order (sorted by name) is preserved.
func stowTreePackageDirs(treeDir string) ([]string, error) {
	entries, err := os.ReadDir(treeDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		empty, err := dirIsEmpty(filepath.Join(treeDir, name))
		if err != nil {
			return nil, err
		}
		if empty {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// dirIsEmpty reports whether dir has no entries at all.
func dirIsEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// pkgFiles returns all non-directory entries (files and symlinks) relative to dir.
func pkgFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}

// pkgMappings derives display mappings: children of each top-level dir in the
// package, e.g. .config/fish → ~/.config/fish.
func pkgMappings(dir string) []string {
	var out []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	home, _ := os.UserHomeDir()
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, filepath.Join(home, e.Name()))
			continue
		}
		children, err := os.ReadDir(filepath.Join(dir, e.Name()))
		if err != nil || len(children) == 0 {
			out = append(out, filepath.Join(home, e.Name()))
			continue
		}
		for _, c := range children {
			out = append(out, filepath.Join(home, e.Name(), c.Name()))
		}
	}
	sort.Strings(out)
	return out
}

// expandTarget expands a leading ~ in the stow target.
func expandTarget(target string) string {
	home, _ := os.UserHomeDir()
	if target == "~" {
		return home
	}
	if strings.HasPrefix(target, "~/") {
		return filepath.Join(home, target[2:])
	}
	return target
}

// stowDir is the repo's stow tree (parent of a package dir): config/. A
// stowed target resolves into stowDir — whether via a direct symlink into
// its own package, a folded parent-dir symlink, or a cross-package symlink
// (e.g. omp's AGENTS.md links into the opencode package). The result is
// symlink-resolved so ownership comparisons stay valid when the path to the
// repo itself contains symlinked components.
func (p *StowPkg) stowDir() string {
	dir := filepath.Dir(p.Dir)
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}

// managed reports whether target path tp is already provided by some stowed
// package (resolves into stowDir). The whole path is resolved — links
// reached through a stow-folded parent directory count as managed — but a
// dangling final symlink is decided from its recorded target instead of
// being treated as absent, so a repo link whose package file moved (or that
// dangles for any other reason) stays recognized as owned and is never
// backed up as a conflict.
func (p *StowPkg) managed(tp string) bool {
	if resolved, err := filepath.EvalSymlinks(tp); err == nil {
		return withinTree(resolved, p.stowDir())
	}
	// Dangling or missing: only a symlink can be repo-owned here.
	fi, err := os.Lstat(tp)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		return false
	}
	dest, err := os.Readlink(tp)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(tp), dest)
	}
	return withinTree(canonicalLeaf(dest), p.stowDir())
}

// withinTree reports whether path equals root or lives underneath it.
func withinTree(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

// canonicalLeaf resolves symlinks in every ancestor of path while leaving a
// possibly-dangling final component untouched.
func canonicalLeaf(path string) string {
	dir, base := filepath.Dir(path), filepath.Base(path)
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(resolved, base)
	}
	return filepath.Clean(path)
}

// targetClass classifies what currently occupies a package file's target.
type targetClass int

const (
	targetAbsent  targetClass = iota // nothing there → stow will create the link
	targetManaged                    // already owned by the repo stow tree
	targetOther                      // real file/dir or foreign symlink → conflict
)

// classifyTarget is the single source of truth shared by conflicts and the
// dry-run plans in configs.go.
func (p *StowPkg) classifyTarget(tp string) targetClass {
	if _, err := os.Lstat(tp); err != nil {
		return targetAbsent
	}
	if p.managed(tp) {
		return targetManaged
	}
	return targetOther
}

// State reports whether the package is fully/partially stowed into target.
func (p *StowPkg) State(target string) StowState {
	stowed := 0
	for _, f := range p.Files {
		tp := filepath.Join(target, f)
		if _, err := os.Lstat(tp); err != nil {
			continue
		}
		if p.managed(tp) {
			stowed++
		}
	}
	switch {
	case stowed == 0:
		return StowNone
	case stowed == len(p.Files):
		return StowFull
	default:
		return StowPartial
	}
}

// conflicts returns package-relative files whose target exists and is not
// already provided by some stowed package. Real files at the target and
// symlinks pointing outside the repo block stow and need backing up.
func (p *StowPkg) conflicts(target string) []string {
	var out []string
	for _, f := range p.Files {
		if p.classifyTarget(filepath.Join(target, f)) == targetOther {
			out = append(out, f)
		}
	}
	return out
}

// backupConflicts moves conflicting target paths into
// ~/.dotfiles-backup/<timestamp>/ preserving relative layout.
func backupConflicts(target string, rels []string) (string, error) {
	if len(rels) == 0 {
		return "", nil
	}
	home, _ := os.UserHomeDir()
	stamp := time.Now().Format("20060102-150405")
	root := filepath.Join(home, ".dotfiles-backup", stamp)
	for _, rel := range rels {
		src := filepath.Join(target, rel)
		dest := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return root, err
		}
		if err := os.Rename(src, dest); err != nil {
			return root, fmt.Errorf("backing up %s: %w", src, err)
		}
	}
	return root, nil
}

// applyStow backs up conflicts, then runs stow --restow for the package.
// Returns the backup dir ("" if none) and error.
func applyStow(repoRoot string, p *StowPkg, target string) (string, error) {
	target = expandTarget(target)
	backup, err := backupConflicts(target, p.conflicts(target))
	if err != nil {
		return backup, err
	}
	cmd := exec.Command("stow", "--restow", "-d", configDir(repoRoot), "-t", target, p.Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return backup, fmt.Errorf("stow %s: %w\n%s", p.Name, err, tail(string(out), 10))
	}
	// Post-stow hook: bat's theme cache must be rebuilt once the themes dir exists.
	if p.Name == "bat" {
		if err := batCacheBuild(); err != nil {
			return backup, err
		}
	}
	return backup, nil
}
