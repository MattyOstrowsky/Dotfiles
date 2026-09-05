package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PreflightStatus is the stable, machine-readable state used by the planner
// and by post-install validation. The values are intentionally plain strings
// so a TUI or CLI can render them without knowing the implementation details.
type PreflightStatus string

const (
	PreflightInstalled   PreflightStatus = "installed"
	PreflightMissing     PreflightStatus = "missing"
	PreflightAvailable   PreflightStatus = "available"
	PreflightUnavailable PreflightStatus = "unavailable"
	PreflightConflict    PreflightStatus = "conflict"
	PreflightUnsupported PreflightStatus = "unsupported"
	PreflightDangling    PreflightStatus = "dangling"

	// Short aliases are useful to callers that do not need to distinguish the
	// preflight and validation phases.
	StatusInstalled   = PreflightInstalled
	StatusMissing     = PreflightMissing
	StatusAvailable   = PreflightAvailable
	StatusUnavailable = PreflightUnavailable
	StatusConflict    = PreflightConflict
	StatusUnsupported = PreflightUnsupported
	StatusDangling    = PreflightDangling
)

// PackagePlan records the state of one native package without installing it.
// It is included in ToolPreflight so dependency availability is visible to a
// caller even though ToolPlan predates per-dependency planning.
type PackagePlan struct {
	Name      string
	Installed bool
	Available bool
	Status    PreflightStatus
	Detail    string
}

// ToolPreflight combines the existing ToolPlan with a status and dependency
// package details suitable for a UI. Plan is always the value returned by
// PlanTools when a manager is available.
type ToolPreflight struct {
	Tool     Tool
	Plan     ToolPlan
	Status   PreflightStatus
	Detail   string
	Packages []PackagePlan
}

// ConfigPlan is the read-only plan for one config package. Config contains the
// host-scope metadata and Stow is the exact per-file plan from PlanStow.
type ConfigPlan struct {
	Config  ConfigPkg
	Package ConfigPkg // compatibility/readability alias for UI consumers
	Stow    *StowPlan
	Plan    *StowPlan // alias for callers that call all plans "Plan"
	Status  PreflightStatus
	Detail  string
	Skipped bool
}

// PreflightReport is a data-only snapshot. BuildPreflight never writes config
// or target files; the only operation that can have side effects is Refresh,
// and that is performed only when refresh is explicitly true.
type PreflightReport struct {
	System           SystemContext
	RepoRoot         string
	Target           string
	Refreshed        bool
	RefreshDone      bool
	RefreshAttempted bool

	ToolPlans   []ToolPlan
	ConfigPlans []ConfigPlan
	// Tools and Configs are convenient aliases for clients that use shorter
	// collection names. They contain the same values as the *Plans fields.
	Tools   []ToolPlan
	Configs []ConfigPlan

	ToolResults   []ToolPreflight
	ConfigResults []ConfigPlan
	Warnings      []string
	Errors        []error
}

// BuildPreflight refreshes package metadata when requested, then computes
// native-tool and config/stow plans. It accepts a caller-selected target; it
// never substitutes the current HOME. A nil manager is supported for
// config-only previews and marks native package tools unsupported.
func BuildPreflight(ctx SystemContext, tools []Tool, pkgs []StowPkg, repoRoot, target string, refresh bool) PreflightReport {
	report := PreflightReport{
		System:   ctx,
		RepoRoot: repoRoot,
		Target:   expandTarget(target),
		Warnings: make([]string, 0),
		Errors:   make([]error, 0),
	}

	if refresh && ctx.Manager != nil {
		report.RefreshAttempted = true
		if err := ctx.Manager.Refresh(); err != nil {
			report.Errors = append(report.Errors, fmt.Errorf("refresh %s package metadata: %w", ctx.Kind(), err))
		} else {
			report.Refreshed = true
			report.RefreshDone = true
		}
	}

	plans := make([]ToolPlan, len(tools))
	if ctx.Manager != nil {
		// PlanTools remains the canonical native-tool planner. Its output is
		// copied below so this report does not expose mutable internal state.
		copy(plans, PlanTools(ctx, tools))
	} else {
		for i := range tools {
			plans[i] = planToolWithoutManager(tools[i])
		}
	}
	report.ToolPlans = plans

	report.ToolResults = make([]ToolPreflight, 0, len(tools))
	for i := range tools {
		p := plans[i]
		normalizeToolPlan(ctx, tools[i], &p)
		plans[i] = p
		result := ToolPreflight{Tool: tools[i], Plan: p}
		result.Packages = packagePlans(ctx, tools[i])
		result.Status, result.Detail = toolPreflightStatus(ctx, tools[i], p, result.Packages)
		report.ToolResults = append(report.ToolResults, result)
		if result.Status == PreflightUnavailable || result.Status == PreflightUnsupported {
			report.Warnings = append(report.Warnings, fmt.Sprintf("tool %s: %s", tools[i].Name, result.Detail))
		}
	}
	report.Tools = append([]ToolPlan(nil), plans...)

	report.ConfigPlans = make([]ConfigPlan, 0, len(pkgs))
	for i := range pkgs {
		p := pkgs[i] // PlanStow only reads the package; keep caller data untouched.
		if p.Dir == "" && repoRoot != "" {
			p.Dir = filepath.Join(configDir(repoRoot), p.Name)
		}
		meta := metaForPkg(p.Name)
		if p.Host != "" {
			meta.Host = p.Host
		}
		stow := PlanStow(&p, report.Target)
		cp := ConfigPlan{Config: meta, Package: meta, Stow: stow, Plan: stow}
		switch {
		case !configTargetSupported(meta.Host, report.Target):
			cp.Status = PreflightUnsupported
			cp.Skipped = true
			cp.Detail = fmt.Sprintf("host scope %s is not valid for Linux target %q", meta.Host, report.Target)
			report.Warnings = append(report.Warnings, fmt.Sprintf("config %s: %s", meta.Name, cp.Detail))
		case stow.HasConflicts():
			cp.Status = PreflightConflict
			cp.Detail = fmt.Sprintf("%d target path(s) would be backed up", len(stow.BackupRels))
			report.Warnings = append(report.Warnings, fmt.Sprintf("config %s: %s", meta.Name, cp.Detail))
		default:
			cp.Status = PreflightAvailable
			if len(stow.Files) == 0 {
				cp.Detail = "config package contains no stow files"
			} else {
				cp.Detail = fmt.Sprintf("%d path(s) ready to stow", len(stow.Files))
			}
		}
		report.ConfigPlans = append(report.ConfigPlans, cp)
	}
	report.Configs = append([]ConfigPlan(nil), report.ConfigPlans...)
	report.ConfigResults = append([]ConfigPlan(nil), report.ConfigPlans...)
	return report
}

func planToolWithoutManager(t Tool) ToolPlan {
	p := ToolPlan{Tool: t.Name, Source: t.Source, Package: t.NativePackage("")}
	_, err := exec.LookPath(t.Bin)
	p.Installed = err == nil
	switch {
	case p.Installed:
		p.Available = true
	case t.Source == sourcePackage:
		p.Reason = "no supported package manager is available"
	case t.Source == sourceNone:
		p.Reason = "external/config-only: binary provided outside the installer"
	default:
		p.Available = true
	}
	return p
}

// normalizeToolPlan makes PlanTools' result authoritative to the supplied
// context. Tool.Installed historically detects the machine on its own, while
// callers (and tests) may provide a manager adapter for another host.
func normalizeToolPlan(ctx SystemContext, t Tool, p *ToolPlan) {
	if ctx.Manager == nil {
		return
	}
	kind := ctx.Kind()
	if t.Source == sourcePackage {
		p.Package = t.NativePackage(kind)
		if p.Package == "" {
			p.Installed = false
			p.Available = false
			p.Reason = fmt.Sprintf("no %s package mapping for %q", kind, t.Name)
			return
		}
		p.Installed = ctx.Manager.Installed(p.Package)
		p.Available = p.Installed || ctx.Manager.Available(p.Package)
		if !p.Available {
			p.Reason = fmt.Sprintf("%s: package %q unavailable", kind, p.Package)
		}
		return
	}
	_, err := exec.LookPath(t.Bin)
	p.Installed = err == nil
	if !p.Installed && t.Source != sourceNone {
		p.Available = true
	}
}

func packagePlans(ctx SystemContext, t Tool) []PackagePlan {
	if ctx.Manager == nil {
		return nil
	}
	kind := ctx.Kind()
	packages := append([]string{t.NativePackage(kind)}, t.NativeDeps(kind)...)
	seen := make(map[string]bool, len(packages))
	out := make([]PackagePlan, 0, len(packages))
	for _, name := range packages {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		installed := ctx.Manager.Installed(name)
		available := installed || ctx.Manager.Available(name)
		status := PreflightUnavailable
		detail := fmt.Sprintf("%s package %q is unavailable", kind, name)
		switch {
		case installed:
			status, detail = PreflightInstalled, "package is installed"
		case available:
			status, detail = PreflightAvailable, "package is available"
		}
		out = append(out, PackagePlan{Name: name, Installed: installed, Available: available, Status: status, Detail: detail})
	}
	return out
}

func toolPreflightStatus(ctx SystemContext, t Tool, p ToolPlan, packages []PackagePlan) (PreflightStatus, string) {
	if p.Installed {
		for _, pkg := range packages {
			if pkg.Status == PreflightUnavailable {
				return PreflightUnavailable, fmt.Sprintf("dependency %q is unavailable", pkg.Name)
			}
		}
		return PreflightInstalled, "tool is already installed"
	}
	if t.Source == sourcePackage {
		if ctx.Manager == nil || !ctx.Kind().Valid() {
			if p.Reason == "" {
				p.Reason = "no supported package manager is available"
			}
			return PreflightUnsupported, p.Reason
		}
		for _, pkg := range packages {
			if pkg.Status == PreflightUnavailable {
				return PreflightUnavailable, pkg.Detail
			}
		}
		if p.Available {
			return PreflightAvailable, "tool package and dependencies are available"
		}
		if p.Reason != "" {
			return PreflightUnavailable, p.Reason
		}
		return PreflightUnavailable, "tool package is unavailable"
	}
	if t.Source == sourceNone {
		return PreflightMissing, "external/config-only binary is not present"
	}
	if p.Available {
		return PreflightAvailable, "tool can be fetched by the installer"
	}
	return PreflightUnsupported, strings.TrimSpace(p.Reason)
}

func configTargetSupported(host HostScope, target string) bool {
	if host != HostWSL {
		return true
	}
	target = expandTarget(target)
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return true
	}
	return target != home && !withinTree(target, home)
}
