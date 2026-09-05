package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed tools.yaml
var toolsYAML []byte

// source values used by tools.yaml. sourcePackage is the legacy "apt" value
// kept for native-package tools — the concrete package per manager comes
// from Tool.Packages (see the tools.yaml header for the schema).
const (
	sourcePackage = "apt"    // native package via the detected package manager
	sourceGitHub  = "github" // release tarball download
	sourceScript  = "script" // remote installer script
	sourceNone    = "none"   // config-only entry, binary external
)

type GitHubSpec struct {
	Repo  string `yaml:"repo"`
	Asset string `yaml:"asset"`
	Kind  string `yaml:"kind"` // tarball-bin | tarball-dir
	re    *regexp.Regexp
}

type Tool struct {
	Name     string `yaml:"name"`
	Bin      string `yaml:"bin"`
	Desc     string `yaml:"desc"`
	Category string `yaml:"category"`
	Source   string `yaml:"source"` // apt (native package) | github | script | none

	// Packages maps a manager kind ("apt","dnf","pacman") to the native
	// package that provides this tool. Deps holds extra per-manager
	// packages needed at runtime or to install from source (build deps).
	// Keys must be known manager kinds; loadManifest rejects anything else.
	Packages map[string]string   `yaml:"packages"`
	Deps     map[string][]string `yaml:"deps"`

	GitHub *GitHubSpec `yaml:"github"`
	Script string      `yaml:"script"`
	Stow   string      `yaml:"stow"`
	Notes  string      `yaml:"notes"`

	// AptPkg/AptDeps mirror Packages["apt"]/Deps["apt"] and keep the
	// legacy apt batch flow (ui.go/main.go) working unchanged.
	AptPkg  string   `yaml:"-"`
	AptDeps []string `yaml:"-"`
}

type Manifest struct {
	Tools []Tool `yaml:"tools"`
}

// loadManifest parses the embedded tools.yaml and validates it: duplicate
// tool names, duplicate stow packages, duplicate native packages per
// manager, required bin fields, known manager keys and per-source
// requirements. Derived AptPkg/AptDeps are filled afterwards.
func loadManifest() (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(toolsYAML, &m); err != nil {
		return nil, fmt.Errorf("parsing embedded tools.yaml: %w", err)
	}
	seenNames := map[string]bool{}
	seenStow := map[string]string{}                // stow dir → first tool claiming it
	seenPkg := map[ManagerKind]map[string]string{} // manager → native package → tool
	for i := range m.Tools {
		t := &m.Tools[i]
		if err := t.validate(seenNames, seenStow, seenPkg); err != nil {
			return nil, err
		}
		t.AptPkg = t.Packages["apt"]
		t.AptDeps = t.Deps["apt"]
	}
	return &m, nil
}

func (t *Tool) validate(seenNames map[string]bool, seenStow map[string]string, seenPkg map[ManagerKind]map[string]string) error {
	if t.Name == "" {
		return fmt.Errorf("tool without a name")
	}
	if seenNames[t.Name] {
		return fmt.Errorf("duplicate tool name %q", t.Name)
	}
	seenNames[t.Name] = true

	if t.Bin == "" {
		return fmt.Errorf("tool %s: missing required bin", t.Name)
	}

	for key, pkg := range t.Packages {
		kind := ManagerKind(key)
		if !kind.Valid() {
			return fmt.Errorf("tool %s: unknown manager kind %q in packages (want apt, dnf or pacman)", t.Name, key)
		}
		if pkg == "" {
			return fmt.Errorf("tool %s: empty package for manager %s", t.Name, kind)
		}
		owners, ok := seenPkg[kind]
		if !ok {
			owners = map[string]string{}
			seenPkg[kind] = owners
		}
		if prev, dup := owners[pkg]; dup {
			return fmt.Errorf("duplicate %s package %q claimed by tools %s and %s", kind, pkg, prev, t.Name)
		}
		owners[pkg] = t.Name
	}

	for key, deps := range t.Deps {
		if kind := ManagerKind(key); !kind.Valid() {
			return fmt.Errorf("tool %s: unknown manager kind %q in deps (want apt, dnf or pacman)", t.Name, key)
		}
		t.Deps[key] = dedupeStrings(deps)
	}

	if t.Stow != "" {
		if prev, dup := seenStow[t.Stow]; dup {
			return fmt.Errorf("duplicate stow package %q used by tools %s and %s", t.Stow, prev, t.Name)
		}
		seenStow[t.Stow] = t.Name
	}

	switch t.Source {
	case sourcePackage:
		if len(t.Packages) == 0 {
			return fmt.Errorf("tool %s: package source requires a per-manager packages map (apt/dnf/pacman)", t.Name)
		}
	case sourceGitHub:
		if t.GitHub == nil || t.GitHub.Repo == "" || t.GitHub.Asset == "" {
			return fmt.Errorf("tool %s: github source requires github.repo and github.asset", t.Name)
		}
		re, err := regexp.Compile(t.GitHub.Asset)
		if err != nil {
			return fmt.Errorf("tool %s: bad asset regex: %w", t.Name, err)
		}
		t.GitHub.re = re
		if t.GitHub.Kind != "tarball-bin" && t.GitHub.Kind != "tarball-dir" {
			return fmt.Errorf("tool %s: unknown github kind %q", t.Name, t.GitHub.Kind)
		}
	case sourceScript:
		if t.Script == "" {
			return fmt.Errorf("tool %s: script source requires script URL", t.Name)
		}
	case sourceNone:
		// config-only entry — the binary lives outside the installer
	default:
		return fmt.Errorf("tool %s: unknown source %q", t.Name, t.Source)
	}
	return nil
}

// dedupeStrings trims, drops empties and deduplicates a package list,
// preserving first-seen order.
func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// NativePackage returns the package providing the tool under the given
// manager kind, or "" when that manager does not provide it.
func (t *Tool) NativePackage(kind ManagerKind) string {
	if p, ok := t.Packages[string(kind)]; ok {
		return p
	}
	if kind == ManagerApt {
		return t.AptPkg // derived fallback (always equals Packages["apt"])
	}
	return ""
}

// NativeDeps returns the extra packages the tool needs under the given
// manager kind (empty when none).
func (t *Tool) NativeDeps(kind ManagerKind) []string {
	if d, ok := t.Deps[string(kind)]; ok {
		return d
	}
	if kind == ManagerApt {
		return t.AptDeps
	}
	return nil
}

// SourceLabel describes where the tool comes from for display.
func (t *Tool) SourceLabel() string {
	switch t.Source {
	case sourcePackage:
		return "apt: " + t.AptPkg
	case sourceGitHub:
		return "github: " + t.GitHub.Repo
	case sourceScript:
		return "script: " + t.Script
	case sourceNone:
		return "external/config-only"
	}
	return t.Source
}

// SourceLabelFor renders a native package using the detected manager.
// Non-native sources keep the generic SourceLabel text.
func (t *Tool) SourceLabelFor(kind ManagerKind) string {
	if t.Source == sourcePackage {
		pkg := t.NativePackage(kind)
		if pkg == "" {
			return string(kind) + ": unavailable"
		}
		return string(kind) + ": " + pkg
	}
	return t.SourceLabel()

}

// DepsLabel returns the full dependency list for display (apt view).
func (t *Tool) DepsLabel() string {
	if len(t.AptDeps) == 0 {
		return "—"
	}
	return strings.Join(t.AptDeps, " ")
}

// Installed reports whether the tool is already present on the host.
// Native-package tools defer to the detected package manager; everything
// else is checked by binary presence on PATH. Only when no supported distro
// is detectable (e.g. a foreign OS) does the legacy dpkg probe run, so apt
// hosts keep today's exact behaviour.
func (t *Tool) Installed() bool {
	if ctx, err := detectSystem(); err == nil {
		if pkg := t.NativePackage(ctx.Kind()); pkg != "" {
			return ctx.Manager.Installed(pkg)
		}
	} else if t.Source == sourcePackage {
		return dpkgInstalled(t.AptPkg)
	}
	_, err := exec.LookPath(t.Bin)
	return err == nil
}

// NativePackageUnion returns the deduplicated union of native packages and
// per-manager deps for the tools, preserving first-seen order (each tool's
// own package before its deps).
func NativePackageUnion(tools []Tool, kind ManagerKind) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for i := range tools {
		add(tools[i].NativePackage(kind))
		for _, d := range tools[i].NativeDeps(kind) {
			add(d)
		}
	}
	return out
}

// AptPackageUnion returns the apt union for the legacy apt batch flow
// (ui.go/main.go). New code should use NativePackageUnion with a detected
// ManagerKind instead.
func AptPackageUnion(tools []Tool) []string {
	return NativePackageUnion(tools, ManagerApt)
}
