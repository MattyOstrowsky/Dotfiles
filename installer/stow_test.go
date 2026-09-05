package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanStowTargetActions(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "config", "example")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "managed"), []byte("from config"), 0o644); err != nil {
		t.Fatal(err)
	}

	newPkg := func(file string) *StowPkg {
		return &StowPkg{Name: "example", Dir: pkgDir, Files: []string{file}}
	}

	t.Run("empty target plans link", func(t *testing.T) {
		target := filepath.Join(root, "empty-target")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}

		plan := PlanStow(newPkg("new-file"), target)
		if len(plan.Files) != 1 || plan.Files[0].Action != PlanLink {
			t.Fatalf("PlanStow() files = %#v, want one link action", plan.Files)
		}
		if plan.Files[0].Managed || plan.HasConflicts() {
			t.Fatalf("empty target plan = %#v, want unmanaged and conflict-free", plan.Files[0])
		}
	})

	t.Run("real file plans backup", func(t *testing.T) {
		target := filepath.Join(root, "conflict-target")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "conflict"), []byte("user data"), 0o644); err != nil {
			t.Fatal(err)
		}

		plan := PlanStow(newPkg("conflict"), target)
		if len(plan.Files) != 1 || plan.Files[0].Action != PlanBackup {
			t.Fatalf("PlanStow() files = %#v, want one backup action", plan.Files)
		}
		if plan.Files[0].Managed || len(plan.BackupRels) != 1 || plan.BackupRels[0] != "conflict" {
			t.Fatalf("conflict plan = %#v, backup rels = %#v", plan.Files[0], plan.BackupRels)
		}
		if plan.BackupDir == "" {
			t.Fatal("conflict plan has empty backup directory")
		}
	})

	t.Run("repo symlink is restowed", func(t *testing.T) {
		target := filepath.Join(root, "managed-target")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(pkgDir, "managed"), filepath.Join(target, "managed")); err != nil {
			t.Fatal(err)
		}

		plan := PlanStow(newPkg("managed"), target)
		if len(plan.Files) != 1 || plan.Files[0].Action != PlanRestow || !plan.Files[0].Managed {
			t.Fatalf("managed symlink plan = %#v, want managed restow", plan.Files)
		}
		if plan.HasConflicts() {
			t.Fatalf("managed symlink was reported as conflict: %#v", plan.BackupRels)
		}
	})

	t.Run("dangling repo symlink is restowed", func(t *testing.T) {
		target := filepath.Join(root, "dangling-target")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		danglingDestination := filepath.Join(pkgDir, "removed-from-config")
		if err := os.Symlink(danglingDestination, filepath.Join(target, "dangling")); err != nil {
			t.Fatal(err)
		}

		plan := PlanStow(newPkg("dangling"), target)
		if len(plan.Files) != 1 || plan.Files[0].Action != PlanRestow || !plan.Files[0].Managed {
			t.Fatalf("dangling repo symlink plan = %#v, want managed restow", plan.Files)
		}
		if plan.HasConflicts() {
			t.Fatalf("dangling repo symlink was reported as conflict: %#v", plan.BackupRels)
		}
	})
}
