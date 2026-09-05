package main

import (
	"reflect"
	"testing"
)

func TestParseOSReleaseResolveManager(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		wantDistro  Distro
		wantManager ManagerKind
		wantErr     bool
	}{
		{
			name: "ubuntu",
			data: "ID=ubuntu\nNAME=\"Ubuntu\"\nPRETTY_NAME=\"Ubuntu 24.04 LTS\"\nVERSION_ID=\"24.04\"\n",
			wantDistro: Distro{
				ID: "ubuntu", Name: "Ubuntu", PrettyName: "Ubuntu 24.04 LTS", VersionID: "24.04",
			},
			wantManager: ManagerApt,
		},
		{
			name: "fedora",
			data: "ID=fedora\nNAME=\"Fedora Linux\"\nPRETTY_NAME=\"Fedora Linux 40\"\nVERSION_ID=\"40\"\n",
			wantDistro: Distro{
				ID: "fedora", Name: "Fedora Linux", PrettyName: "Fedora Linux 40", VersionID: "40",
			},
			wantManager: ManagerDnf,
		},
		{
			name: "arch",
			data: "ID=arch\nNAME=Arch Linux\nPRETTY_NAME=\"Arch Linux\"\n",
			wantDistro: Distro{
				ID: "arch", Name: "Arch Linux", PrettyName: "Arch Linux",
			},
			wantManager: ManagerPacman,
		},
		{
			name: "id like",
			data: "ID=linuxmint\nNAME=\"Linux Mint\"\nID_LIKE=\"ubuntu debian\"\n",
			wantDistro: Distro{
				ID: "linuxmint", Name: "Linux Mint", IDLike: []string{"ubuntu", "debian"},
			},
			wantManager: ManagerApt,
		},
		{
			name:       "unsupported",
			data:       "ID=alpine\nNAME=\"Alpine Linux\"\nPRETTY_NAME=\"Alpine Linux\"\n",
			wantDistro: Distro{ID: "alpine", Name: "Alpine Linux", PrettyName: "Alpine Linux"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDistro, err := parseOSRelease([]byte(tt.data))
			if err != nil {
				t.Fatalf("parseOSRelease() error = %v", err)
			}
			if !reflect.DeepEqual(gotDistro, tt.wantDistro) {
				t.Fatalf("parseOSRelease() = %#v, want %#v", gotDistro, tt.wantDistro)
			}

			gotManager, resolveErr := ResolveManager(gotDistro)
			if tt.wantErr {
				if resolveErr == nil {
					t.Fatalf("ResolveManager() error = nil, want unsupported distro error")
				}
				return
			}
			if resolveErr != nil {
				t.Fatalf("ResolveManager() error = %v", resolveErr)
			}
			if gotManager != tt.wantManager {
				t.Fatalf("ResolveManager() = %q, want %q", gotManager, tt.wantManager)
			}
		})
	}
}
