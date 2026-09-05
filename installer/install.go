package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Minute}

func localBin() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "bin")
}

func localOpt() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "opt")
}

// aptInstall is retained for the TUI compatibility surface; it dispatches to
// the detected native manager rather than invoking apt on every distro.
func aptInstall(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	system, err := detectSystem()
	if err != nil {
		return err
	}
	if err := system.Manager.Install(pkgs); err != nil {
		return fmt.Errorf("%s install failed: %w", system.Kind(), err)
	}
	return nil
}

// aptUpdate is retained for the TUI compatibility surface and refreshes the
// detected manager's metadata (dnf makecache or pacman -Sy as appropriate).
func aptUpdate() error {
	system, err := detectSystem()
	if err != nil {
		return err
	}
	if err := system.Manager.Refresh(); err != nil {
		return fmt.Errorf("%s metadata refresh failed: %w", system.Kind(), err)
	}
	return nil
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// installTool dispatches on the tool source. apt tools are expected to be
// covered by the batch aptInstall; this handles github/script and post-steps.
func installTool(t *Tool) error {
	switch t.Source {
	case "github":
		return installFromGitHub(t)
	case "script":
		return installFromScript(t)
	case "apt":
		return nil // already installed via aptInstall batch
	case "none":
		return nil // config-only entry
	}
	return fmt.Errorf("unknown source %q", t.Source)
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func installFromGitHub(t *Tool) error {
	rel, err := fetchLatestRelease(t.GitHub.Repo)
	if err != nil {
		return err
	}
	var url string
	for _, a := range rel.Assets {
		if t.GitHub.re.MatchString(a.Name) {
			url = a.BrowserDownloadURL
			break
		}
	}
	if url == "" {
		return fmt.Errorf("no asset matching %q in %s (tag %s)", t.GitHub.Asset, t.GitHub.Repo, rel.TagName)
	}

	tmp, err := os.MkdirTemp("", "dl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	archive := filepath.Join(tmp, "asset")
	if err := download(url, archive); err != nil {
		return err
	}

	switch t.GitHub.Kind {
	case "tarball-bin":
		return extractBinary(archive, t.Bin, filepath.Join(localBin(), t.Bin))
	case "tarball-dir":
		return extractAppDir(archive, t.Bin)
	}
	return fmt.Errorf("unknown kind %q", t.GitHub.Kind)
}

func fetchLatestRelease(repo string) (*ghRelease, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dotfiles-install")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s latest release: %w", repo, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github API %s: HTTP %d", repo, resp.StatusCode)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decoding %s release: %w", repo, err)
	}
	return &rel, nil
}

func download(url, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "dotfiles-install")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("downloading %s: HTTP %d", url, resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}

// extractBinary finds the tar entry whose basename == bin and installs it.
func extractBinary(archive, bin, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("reading %s: %w", archive, err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if path.Base(hdr.Name) != bin {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("no entry named %q in archive", bin)
}

// extractAppDir unpacks a tarball whose single top-level directory contains
// the app (nvim), moves it to ~/.local/opt/<topdir> and symlinks the binary.
// The temp dir lives inside ~/.local/opt so the final rename never crosses
// filesystems (/tmp may be tmpfs → EXDEV).
func extractAppDir(archive, bin string) error {
	if err := os.MkdirAll(localOpt(), 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(localOpt(), ".extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	topDir := ""
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// sanitize: no absolute paths, no ..
		name := filepath.Clean(hdr.Name)
		if strings.HasPrefix(name, "..") || filepath.IsAbs(name) {
			return fmt.Errorf("unsafe path in archive: %q", hdr.Name)
		}
		first := strings.SplitN(name, string(filepath.Separator), 2)[0]
		if topDir == "" {
			topDir = first
		}
		dest := filepath.Join(tmp, name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o755|0o500)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			os.Remove(dest)
			if err := os.Symlink(hdr.Linkname, dest); err != nil {
				return err
			}
		}
	}
	if topDir == "" {
		return fmt.Errorf("empty archive")
	}

	optDest := filepath.Join(localOpt(), topDir)
	os.RemoveAll(optDest)
	if err := os.Rename(filepath.Join(tmp, topDir), optDest); err != nil {
		return fmt.Errorf("installing to %s: %w", optDest, err)
	}

	binPath := filepath.Join(optDest, "bin", bin)
	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("expected binary %s not found after extract", binPath)
	}
	link := filepath.Join(localBin(), bin)
	if err := os.MkdirAll(localBin(), 0o755); err != nil {
		return err
	}
	os.Remove(link)
	return os.Symlink(binPath, link)
}

func installFromScript(t *Tool) error {
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("curl -fsSL %s | bash -s -- --no-modify-path", t.Script))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("install script failed: %w\n%s", err, tail(string(out), 20))
	}
	// opencode installs to ~/.opencode/bin — link into ~/.local/bin
	home, _ := os.UserHomeDir()
	real := filepath.Join(home, ".opencode", "bin", t.Bin)
	if _, err := os.Stat(real); err != nil {
		return fmt.Errorf("expected %s after script run: %w", real, err)
	}
	link := filepath.Join(localBin(), t.Bin)
	if err := os.MkdirAll(localBin(), 0o755); err != nil {
		return err
	}
	os.Remove(link)
	return os.Symlink(real, link)
}

// batPostSteps links Debian's batcat to bat only when no bat binary exists,
// then rebuilds the theme cache. Fedora and Arch already provide bat and must
// never receive a dangling /usr/bin/batcat symlink.
func batPostSteps() error {
	bat, batErr := exec.LookPath("bat")
	if batErr != nil {
		bat, batErr = exec.LookPath("batcat")
		if batErr != nil {
			return nil // bat not installed; nothing to do
		}
		link := filepath.Join(localBin(), "bat")
		if err := os.MkdirAll(localBin(), 0o755); err != nil {
			return err
		}
		os.Remove(link)
		if err := os.Symlink(bat, link); err != nil {
			return err
		}
	}
	return batCacheBuild()
}

// batCacheBuild rebuilds bat's theme/syntax cache when bat is available.
// Needed again after the bat package is stowed (themes dir appears then).
func batCacheBuild() error {
	bat := filepath.Join(localBin(), "bat")
	if _, err := os.Stat(bat); err != nil {
		if p, err2 := exec.LookPath("bat"); err2 == nil {
			bat = p
		} else if p, err2 := exec.LookPath("batcat"); err2 == nil {
			bat = p
		} else {
			return nil // bat not installed; nothing to do
		}
	}
	cmd := exec.Command(bat, "cache", "--build")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bat cache --build: %w\n%s", err, tail(string(out), 5))
	}
	return nil
}

// chshFish sets fish as the login shell for the current user.
func chshFish() error {
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("USER not set")
	}
	fish, err := exec.LookPath("fish")
	if err != nil {
		return fmt.Errorf("fish not found on PATH: %w", err)
	}
	cmd := exec.Command("sudo", "-n", "chsh", "-s", fish, user)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chsh failed: %w\n%s", err, tail(string(out), 5))
	}
	return nil
}
