package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakePlannerPackageManager struct {
	kind           ManagerKind
	refreshCalls   int
	installedCalls int
	availableCalls int
}

func (m *fakePlannerPackageManager) Kind() ManagerKind { return m.kind }

func (m *fakePlannerPackageManager) Refresh() error {
	m.refreshCalls++
	return errors.New("refresh should not be called")
}

func (m *fakePlannerPackageManager) Installed(string) bool {
	m.installedCalls++
	return false
}

func (m *fakePlannerPackageManager) Available(string) bool {
	m.availableCalls++
	return false
}

func (m *fakePlannerPackageManager) Install([]string) error { return nil }

func TestBuildPreflightWithoutRefreshReportsUnavailableAndConflict(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "app.conf"), []byte("user config"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager := &fakePlannerPackageManager{kind: ManagerDnf}
	ctx := SystemContext{
		Distro:  Distro{ID: "fedora"},
		Arch:    "x86_64",
		Manager: manager,
	}
	tool := Tool{
		Name:   "missing-native-tool",
		Bin:    filepath.Join(root, "not-installed"),
		Source: sourcePackage,
		Deps:   map[string][]string{string(ManagerDnf): {"missing-native-dependency"}},
	}
	pkg := StowPkg{
		Name: "example",
		Dir:  filepath.Join(root, "config", "example"),
		Files: []string{
			"app.conf",
		},
	}

	report := BuildPreflight(ctx, []Tool{tool}, []StowPkg{pkg}, root, target, false)

	if manager.refreshCalls != 0 {
		t.Fatalf("BuildPreflight(refresh=false) called Refresh %d time(s)", manager.refreshCalls)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("BuildPreflight(refresh=false) errors = %v", report.Errors)
	}
	if len(report.ToolResults) != 1 {
		t.Fatalf("ToolResults length = %d, want 1", len(report.ToolResults))
	}
	if got := report.ToolResults[0].Status; got != PreflightUnavailable {
		t.Fatalf("tool status = %q, want %q", got, PreflightUnavailable)
	}
	if len(report.ConfigPlans) != 1 {
		t.Fatalf("ConfigPlans length = %d, want 1", len(report.ConfigPlans))
	}
	if got := report.ConfigPlans[0].Status; got != PreflightConflict {
		t.Fatalf("config status = %q, want %q", got, PreflightConflict)
	}
	if !report.ConfigPlans[0].Stow.HasConflicts() {
		t.Fatal("config conflict status has no planned backup paths")
	}
	if manager.availableCalls != 1 {
		t.Fatalf("BuildPreflight() called Available %d time(s), want one dependency lookup", manager.availableCalls)
	}
}
