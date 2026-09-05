package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ValidationStatus is the same stable status vocabulary used by preflight.
type ValidationStatus = PreflightStatus

const (
	ValidationInstalled   = PreflightInstalled
	ValidationMissing     = PreflightMissing
	ValidationAvailable   = PreflightAvailable
	ValidationUnavailable = PreflightUnavailable
	ValidationConflict    = PreflightConflict
	ValidationUnsupported = PreflightUnsupported
	ValidationDangling    = PreflightDangling
)

// ToolValidation reports both checks needed for a native tool: its package
// manager record and the executable visible on PATH. For fetched tools only
// BinaryPresent is applicable.
type ToolValidation struct {
	Tool           string
	Binary         string
	Package        string
	BinaryPresent  bool
	PackagePresent bool
	Status         ValidationStatus
	Detail         string
}

// LinkValidation is one config package file's observed target state.
type LinkValidation struct {
	Rel      string
	Target   string
	Status   ValidationStatus
	Detail   string
	Managed  bool
	Symlink  bool
	Dangling bool
}

// ConfigValidation reports every expected stow link and an aggregate status.
type ConfigValidation struct {
	Config ConfigPkg
	Pkg    StowPkg
	Status ValidationStatus
	Detail string
	Files  []LinkValidation
}

// ValidationReport is a data-only post-install result. No manager is needed
// when validating config-only inputs (and a nil manager is accepted).
type ValidationReport struct {
	Target   string
	Tools    []ToolValidation
	Configs  []ConfigValidation
	Valid    bool
	Warnings []string
	Errors   []error
}

// ValidateInstallation checks binaries/packages and all expected stow links
// under target. It does not install, refresh, repair, or otherwise mutate the
// filesystem. A nil PackageManager is valid for config-only validation.
func ValidateInstallation(ctx SystemContext, tools []Tool, pkgs []StowPkg, target string) ValidationReport {
	target = expandTarget(target)
	report := ValidationReport{
		Target:   target,
		Tools:    make([]ToolValidation, 0, len(tools)),
		Configs:  make([]ConfigValidation, 0, len(pkgs)),
		Warnings: make([]string, 0),
		Errors:   make([]error, 0),
		Valid:    true,
	}
	for i := range tools {
		result := validateTool(ctx, tools[i])
		report.Tools = append(report.Tools, result)
		if result.Status != ValidationInstalled {
			report.Valid = false
			report.Warnings = append(report.Warnings, fmt.Sprintf("tool %s: %s", result.Tool, result.Detail))
		}
	}
	for i := range pkgs {
		result := validateConfig(pkgs[i], target)
		report.Configs = append(report.Configs, result)
		if result.Status != ValidationInstalled {
			report.Valid = false
			report.Warnings = append(report.Warnings, fmt.Sprintf("config %s: %s", result.Config.Name, result.Detail))
		}
	}
	return report
}

// ValidatePostInstall is a descriptive alias for ValidateInstallation.
func ValidatePostInstall(ctx SystemContext, tools []Tool, pkgs []StowPkg, target string) ValidationReport {
	return ValidateInstallation(ctx, tools, pkgs, target)
}

// ValidatePreflight is retained as a convenient phase-neutral name for CLI/TUI
// integrations that validate immediately after executing a preflight plan.
func ValidatePreflight(ctx SystemContext, tools []Tool, pkgs []StowPkg, target string) ValidationReport {
	return ValidateInstallation(ctx, tools, pkgs, target)
}

func validateTool(ctx SystemContext, t Tool) ToolValidation {
	result := ToolValidation{Tool: t.Name, Binary: t.Bin, Status: ValidationMissing}
	if t.Bin != "" {
		_, err := exec.LookPath(t.Bin)
		result.BinaryPresent = err == nil
	}
	if t.Source == sourcePackage {
		if ctx.Manager == nil || !ctx.Kind().Valid() {
			result.Status = ValidationUnsupported
			result.Detail = "cannot verify native package without a supported package manager"
			return result
		}
		result.Package = t.NativePackage(ctx.Kind())
		if result.Package == "" {
			result.Status = ValidationUnsupported
			result.Detail = fmt.Sprintf("no %s package mapping", ctx.Kind())
			return result
		}
		result.PackagePresent = ctx.Manager.Installed(result.Package)
		switch {
		case result.PackagePresent && result.BinaryPresent:
			result.Status, result.Detail = ValidationInstalled, "package and binary are present"
		case result.PackagePresent:
			result.Detail = fmt.Sprintf("package %q is installed but binary %q is missing", result.Package, t.Bin)
		case result.BinaryPresent:
			result.Detail = fmt.Sprintf("binary %q is present but package %q is not installed", t.Bin, result.Package)
		default:
			result.Detail = fmt.Sprintf("package %q and binary %q are missing", result.Package, t.Bin)
		}
		return result
	}
	if result.BinaryPresent {
		result.Status, result.Detail = ValidationInstalled, "binary is present"
	} else if t.Source == sourceNone {
		result.Status = ValidationMissing
		result.Detail = fmt.Sprintf("external binary %q is missing", t.Bin)
	} else {
		result.Detail = fmt.Sprintf("binary %q is missing", t.Bin)
	}
	return result
}

func validateConfig(p StowPkg, target string) ConfigValidation {
	meta := metaForPkg(p.Name)
	if p.Host != "" {
		meta.Host = p.Host
	}
	result := ConfigValidation{Config: meta, Pkg: p, Status: ValidationInstalled, Files: make([]LinkValidation, 0, len(p.Files))}
	if !configTargetSupported(meta.Host, target) {
		result.Status = ValidationUnsupported
		result.Detail = fmt.Sprintf("host scope %s is not valid for Linux target %q", meta.Host, target)
		return result
	}

	missing, conflict, dangling := 0, 0, 0
	for _, rel := range p.Files {
		tp := filepath.Join(target, rel)
		link := LinkValidation{Rel: rel, Target: tp, Status: ValidationMissing}
		fi, err := os.Lstat(tp)
		if err != nil {
			if !os.IsNotExist(err) {
				link.Detail = err.Error()
			}
			missing++
			result.Files = append(result.Files, link)
			continue
		}
		link.Symlink = fi.Mode()&os.ModeSymlink != 0
		if link.Symlink {
			if _, err := os.Stat(tp); err != nil {
				link.Status = ValidationDangling
				link.Dangling = true
				link.Detail = "symlink target does not exist"
				dangling++
				result.Files = append(result.Files, link)
				continue
			}
		}
		link.Managed = p.managed(tp)
		if link.Managed {
			link.Status = ValidationInstalled
			link.Detail = "stow-managed link"
		} else {
			link.Status = ValidationConflict
			link.Detail = "target exists but is not managed by this config tree"
			conflict++
		}
		result.Files = append(result.Files, link)
	}

	switch {
	case dangling > 0:
		result.Status = ValidationDangling
		result.Detail = fmt.Sprintf("%d dangling stow link(s)", dangling)
	case conflict > 0:
		result.Status = ValidationConflict
		result.Detail = fmt.Sprintf("%d target path(s) are not stow-managed", conflict)
	case missing > 0:
		result.Status = ValidationMissing
		result.Detail = fmt.Sprintf("%d stow link(s) are missing", missing)
	case len(p.Files) == 0:
		result.Status = ValidationMissing
		result.Detail = "config package contains no stow files"
	default:
		result.Detail = fmt.Sprintf("%d stow link(s) are present", len(p.Files))
	}
	return result
}

// ValidateConfigs is a config-only convenience API; it deliberately does not
// require or inspect a PackageManager.
func ValidateConfigs(pkgs []StowPkg, target string) ValidationReport {
	return ValidateInstallation(SystemContext{}, nil, pkgs, target)
}
