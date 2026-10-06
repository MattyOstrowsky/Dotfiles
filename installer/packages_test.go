package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerDetection(t *testing.T) {
	for data, want := range map[string]string{"ID=ubuntu\nVERSION_ID=26.04": "apt", "ID=linuxmint\nID_LIKE=\"ubuntu debian\"": "apt", "ID=fedora": "dnf", "ID=arch": "pacman"} {
		got, err := managerFromRelease(data)
		if err != nil || got != want {
			t.Fatal(got, err)
		}
	}
	if _, err := managerFromRelease("ID=unknown"); err == nil {
		t.Fatal("unsupported distro")
	}
}
func TestNativeDryRunNeverExecutesPackageManager(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	script := "#!/bin/sh\ncase \"$3\" in git) printf installed;; *) exit 1;; esac\n"
	os.WriteFile(filepath.Join(dir, "dpkg-query"), []byte(script), 0o755)
	marker := filepath.Join(dir, "ran")
	os.WriteFile(filepath.Join(dir, "apt-get"), []byte("#!/bin/sh\ntouch "+marker+"\nexit 99\n"), 0o755)
	p := defaultProfile()
	p.Configs = false
	o := Options{Target: dir, BinDir: filepath.Join(dir, "bin"), DryRun: true}
	r := Result{}
	tools := []Tool{{Name: "git", Packages: map[string][]string{"apt": {"git", "missing-package"}}}}
	if manager, _ := detectManager(); manager != "apt" {
		t.Skip("apt host required for integration of detection")
	}
	if err := apply(context.Background(), tools, p, o, &r); err != nil {
		t.Fatal(err)
	}
	if !r.Changed {
		t.Fatal("missing package wasn't planned")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("executed apt-get")
	}
}
