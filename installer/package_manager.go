package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// PackageManager is the narrow adapter a planner/executor needs to deal
// with a native package manager. Implementations never install anything
// during construction or detection; Installed/Available only query local
// state and locally cached repository metadata, while Refresh and Install
// are the only privileged operations (run via `sudo -n`).
//
// No AUR/paru/yay or other user-repository helper is ever invoked: pacman
// targets are official Arch repositories only.
type PackageManager interface {
	// Kind identifies the manager ("apt", "dnf", "pacman").
	Kind() ManagerKind
	// Refresh updates local repository metadata (apt-get update / dnf
	// makecache / pacman -Sy). It does not upgrade installed packages.
	Refresh() error
	// Installed reports whether pkg is present locally.
	Installed(pkg string) bool
	// Available reports whether pkg is known to the repositories (or
	// already installed) — an offline, cache-only lookup.
	Available(pkg string) bool
	// Install installs pkgs from the native repositories. pkgs may be
	// empty, in which case Install is a no-op.
	Install(pkgs []string) error
}

// NewPackageManager returns the adapter for a supported manager kind.
// Construction never executes commands and never installs anything.
func NewPackageManager(kind ManagerKind) (PackageManager, error) {
	switch kind {
	case ManagerApt:
		return aptManager{}, nil
	case ManagerDnf:
		return dnfManager{}, nil
	case ManagerPacman:
		return pacmanManager{}, nil
	}
	return nil, fmt.Errorf("unsupported package manager %q (supported: apt, dnf, pacman)", kind)
}

// --- shared exec helpers ------------------------------------------------

// sudoExec runs args under `sudo -n` (non-interactive; credentials must be
// validated beforehand, e.g. by preflight) and folds output into the error.
func sudoExec(args ...string) error {
	cmd := exec.Command("sudo", append([]string{"-n"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w\n%s", args[0], err, tail(string(out), 10))
	}
	return nil
}

// plainExecOK reports whether a command exits 0 (queries only — no sudo).
func plainExecOK(args ...string) bool {
	return exec.Command(args[0], args[1:]...).Run() == nil
}

// dpkgInstalled reports whether a dpkg-managed package is in the
// "install ok installed" state (the apt adapter's local-state probe).
func dpkgInstalled(pkg string) bool {
	out, err := exec.Command("dpkg-query", "-W", "-f=${Status}", pkg).Output()
	return err == nil && strings.Contains(string(out), "install ok installed")
}

// --- apt (Debian/Ubuntu) -----------------------------------------------

type aptManager struct{}

func (aptManager) Kind() ManagerKind { return ManagerApt }

func (aptManager) Refresh() error {
	// Mirrors the historical aptUpdate used by the legacy flow.
	return sudoExec("apt-get", "update")
}

func (aptManager) Installed(pkg string) bool { return dpkgInstalled(pkg) }

func (aptManager) Available(pkg string) bool {
	// apt-cache show lists known packages (installed or not) from the
	// local package lists; unknown packages print nothing.
	out, err := exec.Command("apt-cache", "show", pkg).Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

func (aptManager) Install(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"apt-get", "install", "-y"}, pkgs...)
	return sudoExec(args...)
}

// --- dnf (Fedora/RHEL-family) -------------------------------------------

type dnfManager struct{}

func (dnfManager) Kind() ManagerKind { return ManagerDnf }

func (dnfManager) Refresh() error {
	return sudoExec("dnf", "-q", "makecache")
}

func (dnfManager) Installed(pkg string) bool {
	// dnf keeps its database in rpm.
	return plainExecOK("rpm", "-q", pkg)
}

func (m dnfManager) Available(pkg string) bool {
	if m.Installed(pkg) {
		return true
	}
	// --cacheonly keeps the probe offline after Refresh; fall back to a
	// plain list on dnf versions without the flag.
	if plainExecOK("dnf", "-q", "--cacheonly", "list", "--available", pkg) {
		return true
	}
	return plainExecOK("dnf", "-q", "list", "--available", pkg)
}

func (dnfManager) Install(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"dnf", "install", "-y"}, pkgs...)
	return sudoExec(args...)
}

// --- pacman (Arch) ------------------------------------------------------

type pacmanManager struct{}

func (pacmanManager) Kind() ManagerKind { return ManagerPacman }

func (pacmanManager) Refresh() error {
	// Metadata-only refresh, mirroring apt-get update / dnf makecache.
	// Deliberately NOT -Syu: upgrading the whole system is the operator's
	// separate decision, and the planner runs Refresh immediately before
	// its own Install batch to avoid a partial-upgrade window.
	return sudoExec("pacman", "-Sy", "--noconfirm")
}

func (pacmanManager) Installed(pkg string) bool {
	return plainExecOK("pacman", "-Q", pkg)
}

func (pacmanManager) Available(pkg string) bool {
	// -Si reads the synced sync databases only; no network, no install.
	return plainExecOK("pacman", "-Si", pkg)
}

func (pacmanManager) Install(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	// Official repositories only — never paru/yay/AUR.
	args := append([]string{"pacman", "-S", "--needed", "--noconfirm"}, pkgs...)
	return sudoExec(args...)
}

// --- planning helpers ---------------------------------------------------

// ToolPlan is one tool's provisioning status on a detected system: which
// native package would provide it (if any), whether it is already present,
// and whether the manager can still install it. Computing a plan performs
// no installs and no privileged commands.
type ToolPlan struct {
	Tool      string // manifest name
	Source    string // manifest source value (apt|github|script|none)
	Package   string // native package under the detected manager ("" = none)
	Installed bool   // already present locally (binary or native package)
	Available bool   // installable from the native repos (or fetchable)
	Reason    string // human note when the tool is not provisionable
}

// PlanTools evaluates tools against ctx for the future planner/UI: local
// state first (Installed), then repository availability for native
// packages. github/script downloads are always fetchable; a native-package
// tool without a mapping for this manager is reported as not available.
func PlanTools(ctx SystemContext, tools []Tool) []ToolPlan {
	kind := ctx.Kind()
	plans := make([]ToolPlan, 0, len(tools))
	for i := range tools {
		t := &tools[i]
		p := ToolPlan{
			Tool:      t.Name,
			Source:    t.Source,
			Package:   t.NativePackage(kind),
			Installed: t.Installed(),
		}
		switch {
		case p.Installed:
			p.Available = true
		case p.Package != "":
			p.Available = ctx.Manager.Available(p.Package)
			if !p.Available {
				p.Reason = fmt.Sprintf("%s: package %q not found in repository metadata (run refresh?)", kind, p.Package)
			}
		case t.Source == sourcePackage:
			p.Reason = fmt.Sprintf("no %s package mapping for %q", kind, t.Name)
		case t.Source == sourceNone:
			p.Reason = "external/config-only: binary provided outside the installer"
		default:
			p.Available = true // github/script sources are fetched, not installed
		}
		plans = append(plans, p)
	}
	return plans
}
