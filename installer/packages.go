package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func detectManager() (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("package installation supports Linux only")
	}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	return managerFromRelease(string(data))
}
func managerFromRelease(data string) (string, error) {
	ids := []string{}
	for _, line := range strings.Split(data, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok && (k == "ID" || k == "ID_LIKE") {
			ids = append(ids, strings.Fields(strings.Trim(v, "\"'"))...)
		}
	}
	for _, id := range ids {
		switch id {
		case "ubuntu", "debian":
			return "apt", nil
		case "fedora", "rhel", "centos":
			return "dnf", nil
		case "arch", "manjaro":
			return "pacman", nil
		}
	}
	return "", fmt.Errorf("unsupported distribution: %v", ids)
}
func packageInstalled(ctx context.Context, manager, pkg string) (bool, error) {
	var cmd *exec.Cmd
	switch manager {
	case "apt":
		cmd = exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${db:Status-Status}", pkg)
	case "dnf":
		cmd = exec.CommandContext(ctx, "rpm", "-q", pkg)
	case "pacman":
		cmd = exec.CommandContext(ctx, "pacman", "-Q", pkg)
	default:
		return false, fmt.Errorf("unsupported manager %s", manager)
	}
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("query package %s: %w", pkg, err)
	}
	return manager != "apt" || strings.TrimSpace(string(out)) == "installed", nil
}
func runCommand(ctx context.Context, privileged bool, name string, args ...string) error {
	if privileged && os.Geteuid() != 0 {
		args = append([]string{"-n", name}, args...)
		name = "sudo"
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
func installPackages(ctx context.Context, manager string, pkgs []string, interactive bool) error {
	if len(pkgs) == 0 {
		return nil
	}
	if interactive && os.Geteuid() != 0 {
		cmd := exec.CommandContext(ctx, "sudo", "-v")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	switch manager {
	case "apt":
		if err := runCommand(ctx, true, "apt-get", "update"); err != nil {
			return err
		}
		return runCommand(ctx, true, "apt-get", append([]string{"install", "-y", "--no-install-recommends"}, pkgs...)...)
	case "dnf":
		return runCommand(ctx, true, "dnf", append([]string{"install", "-y", "--setopt=install_weak_deps=False"}, pkgs...)...)
	case "pacman":
		return runCommand(ctx, true, "pacman", append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)...)
	}
	return fmt.Errorf("unsupported package manager %s", manager)
}
func binaryExists(bin, binDir string) bool {
	if info, err := os.Stat(filepath.Join(binDir, bin)); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
		return true
	}
	_, err := exec.LookPath(bin)
	return err == nil
}
func apply(ctx context.Context, tools []Tool, p Profile, o Options, r *Result) error {
	// Validate all selected config destinations and package mappings first.
	configs, err := planConfigs(tools, p, o)
	if err != nil {
		return err
	}
	manager := ""
	missing := []string{}
	releases := []Tool{}
	if p.Install {
		manager, err = detectManager()
		if err != nil {
			return err
		}
		packages := map[string]bool{}
		for _, pkg := range p.Packages {
			packages[pkg] = true
		}
		for _, t := range tools {
			if t.External {
				r.Actions = append(r.Actions, Action{"external", t.Name, "unmanaged"})
				continue
			}
			if t.Release != nil {
				if t.Release.Assets[runtime.GOARCH] == "" && (t.Release.URL == "" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64")) {
					return fmt.Errorf("%s: no release for architecture %s", t.Name, runtime.GOARCH)
				}
				if binaryExists(t.Bin, o.BinDir) {
					r.Actions = append(r.Actions, Action{"binary", t.Name, "ok"})
				} else {
					releases = append(releases, t)
				}
				continue
			}
			pkgs := t.Packages[manager]
			if len(pkgs) == 0 {
				return fmt.Errorf("%s has no %s package mapping; install it externally or choose --configs-only", t.Name, manager)
			}
			for _, pkg := range pkgs {
				packages[pkg] = true
			}
		}
		names := []string{}
		for pkg := range packages {
			names = append(names, pkg)
		}
		sort.Strings(names)
		for _, pkg := range names {
			installed, e := packageInstalled(ctx, manager, pkg)
			if e != nil {
				return e
			}
			if installed {
				r.Actions = append(r.Actions, Action{"package", pkg, "ok"})
			} else {
				missing = append(missing, pkg)
			}
		}
	}
	aliases := ConfigPlan{}
	if p.Install && manager == "apt" {
		for _, t := range tools {
			aliasSource := ""
			switch t.Name {
			case "bat":
				aliasSource = "/usr/bin/batcat"
			case "fd-find":
				aliasSource = "/usr/bin/fdfind"
			}
			if aliasSource != "" && !binaryExists(t.Bin, o.BinDir) {
				aliases.Files = append(aliases.Files, ConfigFile{Source: aliasSource, Dest: filepath.Join(o.BinDir, t.Bin), Change: true})
			}
		}
	}
	if o.DryRun {
		for _, pkg := range missing {
			r.Actions = append(r.Actions, Action{"package", pkg, "planned"})
		}
		for _, t := range releases {
			r.Actions = append(r.Actions, Action{"binary", t.Name, "planned"})
		}
		for _, dir := range configs.Detach {
			r.Actions = append(r.Actions, Action{"detach", dir, "planned"})
		}
		r.Changed = len(missing)+len(releases)+len(configs.Detach)+len(aliases.Files) > 0
		for _, f := range aliases.Files {
			r.Actions = append(r.Actions, Action{"binary-link", f.Dest, "planned"})
		}
		for _, f := range configs.Files {
			s := "ok"
			if f.Change {
				s = "planned"
				r.Changed = true
			}
			r.Actions = append(r.Actions, Action{"config", f.Dest, s})
		}
		return nil
	}
	// Download and verify releases before installing native packages.
	staged := map[string]string{}
	if len(releases) > 0 {
		tmp, e := os.MkdirTemp("", "dotfiles-download-*")
		if e != nil {
			return e
		}
		defer os.RemoveAll(tmp)
		for _, t := range releases {
			path, e := stageRelease(ctx, t, tmp)
			if e != nil {
				return e
			}
			staged[t.Name] = path
		}
	}
	if len(missing) > 0 {
		installErr := installPackages(ctx, manager, missing, o.Interactive)
		// Query after failures too: package managers can partially succeed.
		for _, pkg := range missing {
			installed, e := packageInstalled(ctx, manager, pkg)
			if e != nil {
				return e
			}
			if installed {
				r.Changed = true
				r.Actions = append(r.Actions, Action{"package", pkg, "changed"})
			} else if installErr == nil {
				return fmt.Errorf("package %s missing after installation", pkg)
			}
		}
		if installErr != nil {
			return installErr
		}
	}
	for _, t := range releases {
		dest := filepath.Join(o.BinDir, t.Bin)
		data, e := os.ReadFile(staged[t.Name])
		if e != nil {
			return e
		}
		if e = applyConfigs(ConfigPlan{Files: []ConfigFile{{Dest: dest, Data: data, Mode: 0o755, Change: true}}}, Profile{ConfigMode: "copy"}, r); e != nil {
			return e
		}
	}
	for _, f := range aliases.Files {
		if _, err = os.Stat(f.Source); err != nil {
			return err
		}
	}
	if err = applyConfigs(aliases, Profile{ConfigMode: "link"}, r); err != nil {
		return err
	}
	if err = applyConfigs(configs, p, r); err != nil {
		return err
	}
	// Nord is built into bat, so no mutable theme cache is required.
	return nil
}
