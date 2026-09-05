package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// ManagerKind identifies a supported native package manager.
type ManagerKind string

const (
	ManagerApt    ManagerKind = "apt"
	ManagerDnf    ManagerKind = "dnf"
	ManagerPacman ManagerKind = "pacman"
)

// supportedManagers is the ordered set of kinds NewPackageManager accepts.
var supportedManagers = []ManagerKind{ManagerApt, ManagerDnf, ManagerPacman}

// String implements fmt.Stringer.
func (k ManagerKind) String() string { return string(k) }

// Valid reports whether k is a supported manager kind.
func (k ManagerKind) Valid() bool {
	switch k {
	case ManagerApt, ManagerDnf, ManagerPacman:
		return true
	}
	return false
}

// Distro is the host distribution summary parsed from /etc/os-release.
type Distro struct {
	ID         string   // ID=ubuntu
	Name       string   // NAME="Ubuntu"
	PrettyName string   // PRETTY_NAME="Ubuntu 24.04.1 LTS"
	VersionID  string   // VERSION_ID="24.04"
	IDLike     []string // ID_LIKE="debian" (space separated in the file)
}

// String renders a short human-readable distro summary.
func (d Distro) String() string {
	switch {
	case d.PrettyName != "":
		return d.PrettyName
	case d.Name != "":
		return d.Name
	case d.ID != "":
		return d.ID
	}
	return "unknown distro"
}

// parseOSRelease parses /etc/os-release KEY=VALUE content. Values may be
// bare, single-quoted or double-quoted; double quotes honour the \n, \t,
// \" and \\ escapes from the os-release(5) spec via strconv.Unquote.
func parseOSRelease(data []byte) (Distro, error) {
	fields := make(map[string]string, 8)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue // malformed line; os-release(5) says to ignore it
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			switch {
			case val[0] == '"' && val[len(val)-1] == '"':
				if unquoted, err := strconv.Unquote(val); err == nil {
					val = unquoted
				} else {
					val = val[1 : len(val)-1]
				}
			case val[0] == '\'' && val[len(val)-1] == '\'':
				val = val[1 : len(val)-1]
			}
		}
		fields[key] = val
	}
	d := Distro{
		ID:         fields["ID"],
		Name:       fields["NAME"],
		PrettyName: fields["PRETTY_NAME"],
		VersionID:  fields["VERSION_ID"],
	}
	if like := fields["ID_LIKE"]; like != "" {
		for _, x := range strings.Fields(like) {
			d.IDLike = append(d.IDLike, x)
		}
	}
	if d.ID == "" {
		return Distro{}, fmt.Errorf("os-release: missing required ID field")
	}
	return d, nil
}

// ResolveManager maps a distro to the package manager it natively uses.
// The ID_LIKE list is honoured so derivatives (Linux Mint, Manjaro, Rocky,
// …) resolve without an explicit entry. Unsupported distros return an
// error; nothing is ever installed or auto-fallback'ed to AUR/paru/yay.
func ResolveManager(d Distro) (ManagerKind, error) {
	id := strings.ToLower(strings.TrimSpace(d.ID))
	like := make(map[string]bool, len(d.IDLike))
	for _, x := range d.IDLike {
		like[strings.ToLower(x)] = true
	}
	switch {
	case id == "debian" || id == "ubuntu" || like["debian"] || like["ubuntu"]:
		return ManagerApt, nil
	case id == "fedora" || like["fedora"] || like["rhel"]:
		// RHEL-family 8+/Fedora derivatives default to dnf; ID_LIKE
		// "rhel"/"fedora" covers Rocky, AlmaLinux, CentOS Stream, …
		return ManagerDnf, nil
	case id == "arch" || like["arch"] || id == "manjaro":
		return ManagerPacman, nil
	}
	return "", unsupportedDistroError(d)
}

func unsupportedDistroError(d Distro) error {
	return fmt.Errorf(
		"unsupported distro %q (ID=%q ID_LIKE=%q): supported are Debian/Ubuntu (apt), Fedora/RHEL-family (dnf) and Arch (pacman); "+
			"refusing to install without a native package manager (no automatic AUR/paru fallback)",
		d.String(), d.ID, strings.Join(d.IDLike, " "))
}

// SystemContext bundles the host facts a planner needs: the distro, a
// normalized architecture and the resolved package manager adapter.
// Construction performs no installation and no privileged commands.
type SystemContext struct {
	Distro  Distro
	Arch    string // normalized: x86_64 | aarch64 | i386 | …
	Manager PackageManager
}

// Kind is a shortcut for Manager.Kind(); it returns "" when Manager is nil.
func (c SystemContext) Kind() ManagerKind {
	if c.Manager == nil {
		return ""
	}
	return c.Manager.Kind()
}

// detectSystem reads /etc/os-release (falling back to /usr/lib/os-release),
// derives the CPU architecture and resolves the native package manager in
// one call. It never installs or modifies anything — pure detection.
func detectSystem() (SystemContext, error) {
	for _, p := range []string{"/etc/os-release", "/usr/lib/os-release"} {
		if _, err := os.Stat(p); err == nil {
			return detectSystemFile(p)
		}
	}
	return SystemContext{}, fmt.Errorf("no os-release file found (checked /etc/os-release and /usr/lib/os-release)")
}

// detectSystemFile is detectSystem with an explicit os-release path, for
// tests/fixtures and non-standard roots.
func detectSystemFile(path string) (SystemContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SystemContext{}, fmt.Errorf("reading %s: %w", path, err)
	}
	return detectSystemData(data)
}

// detectSystemData builds a SystemContext from raw os-release content.
func detectSystemData(data []byte) (SystemContext, error) {
	distro, err := parseOSRelease(data)
	if err != nil {
		return SystemContext{}, err
	}
	kind, err := ResolveManager(distro)
	if err != nil {
		return SystemContext{}, err
	}
	mgr, err := NewPackageManager(kind)
	if err != nil {
		return SystemContext{}, err
	}
	return SystemContext{Distro: distro, Arch: linuxArch(runtime.GOARCH), Manager: mgr}, nil
}

// linuxArch maps runtime.GOARCH onto the release-asset naming convention
// used across the github tarballs (x86_64/aarch64), so later asset
// selection can key off the same strings.
func linuxArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i386"
	}
	return goarch
}
