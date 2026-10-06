package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var releaseClient = &http.Client{Timeout: 3 * time.Minute}

type releaseAsset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	Digest string `json:"digest"`
}

func fetch(ctx context.Context, url string, limit int64) ([]byte, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("HTTPS required: %s", url)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "dotfiles-install")
	resp, err := releaseClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("download exceeds %d bytes", limit)
	}
	return data, nil
}
func stageRelease(ctx context.Context, t Tool, dir string) (string, error) {
	fmt.Fprintln(os.Stderr, "Downloading and verifying", t.Name)
	if t.Release.URL != "" {
		return stageDirectRelease(ctx, t, dir)
	}
	data, err := fetch(ctx, "https://api.github.com/repos/"+t.Release.Repo+"/releases/latest", 4<<20)
	if err != nil {
		return "", err
	}
	var release struct {
		Assets []releaseAsset `json:"assets"`
	}
	if err = json.Unmarshal(data, &release); err != nil {
		return "", err
	}
	pattern := regexp.MustCompile(t.Release.Assets[runtime.GOARCH])
	var matches []releaseAsset
	for _, a := range release.Assets {
		if pattern.MatchString(a.Name) {
			matches = append(matches, a)
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("%s: expected one %s asset, found %d", t.Name, runtime.GOARCH, len(matches))
	}
	asset := matches[0]
	expected := strings.TrimPrefix(asset.Digest, "sha256:")
	if !strings.HasPrefix(asset.Digest, "sha256:") {
		expected = ""
		for _, a := range release.Assets {
			lower := strings.ToLower(a.Name)
			if strings.Contains(lower, "checksum") || strings.HasSuffix(lower, "sha256sum") || strings.HasSuffix(lower, "sha256") {
				sums, e := fetch(ctx, a.URL, 4<<20)
				if e != nil {
					return "", e
				}
				for _, line := range strings.Split(string(sums), "\n") {
					fields := strings.Fields(line)
					if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == asset.Name {
						expected = fields[0]
					}
				}
			}
		}
	}
	if hash, e := hex.DecodeString(expected); e != nil || len(hash) != 32 {
		return "", fmt.Errorf("%s: no SHA-256 digest/checksum for %s", t.Name, asset.Name)
	}
	data, err = fetch(ctx, asset.URL, 128<<20)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	if !strings.EqualFold(hex.EncodeToString(sum[:]), expected) {
		return "", fmt.Errorf("%s: SHA-256 mismatch", t.Name)
	}
	archive := filepath.Join(dir, t.Name+".tar.gz")
	if err = os.WriteFile(archive, data, 0o600); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, t.Bin)
	if err = extractBinary(archive, t.Bin, dest); err != nil {
		return "", err
	}
	return dest, nil
}
func extractBinary(archive, bin, dest string) error {
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
	for {
		header, e := tr.Next()
		if e == io.EOF {
			return fmt.Errorf("binary %s missing from archive", bin)
		}
		if e != nil {
			return e
		}
		// Only a regular file is extracted to our chosen filename. Archive paths
		// and symlinks never control the destination.
		if header.Typeflag != tar.TypeReg || path.Base(header.Name) != bin {
			continue
		}
		if header.Size > 256<<20 {
			return fmt.Errorf("binary is too large")
		}
		out, e := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
		if e != nil {
			return e
		}
		_, e = io.Copy(out, tr)
		closeErr := out.Close()
		if e != nil {
			return e
		}
		return closeErr
	}
}

// kubectl and Helm publish their release checksums outside GitHub assets.
func stageDirectRelease(ctx context.Context, t Tool, dir string) (string, error) {
	versionURL := t.Release.VersionURL
	if versionURL == "" {
		versionURL = "https://api.github.com/repos/" + t.Release.Repo + "/releases/latest"
	}
	data, err := fetch(ctx, versionURL, 4<<20)
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(data))
	if t.Release.VersionURL == "" {
		var rel struct {
			Tag string `json:"tag_name"`
		}
		if err = json.Unmarshal(data, &rel); err != nil {
			return "", err
		}
		version = rel.Tag
	}
	if !regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z0-9.-]*$`).MatchString(version) {
		return "", fmt.Errorf("invalid upstream version %q", version)
	}
	replace := strings.NewReplacer("{version}", version, "{arch}", runtime.GOARCH)
	sums, err := fetch(ctx, replace.Replace(t.Release.ChecksumURL), 1<<20)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(sums))
	if len(fields) == 0 {
		return "", fmt.Errorf("missing checksum")
	}
	expected, err := hex.DecodeString(fields[0])
	if err != nil || len(expected) != 32 {
		return "", fmt.Errorf("invalid checksum")
	}
	data, err = fetch(ctx, replace.Replace(t.Release.URL), 128<<20)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	if !bytes.Equal(hash[:], expected) {
		return "", fmt.Errorf("%s: SHA-256 mismatch", t.Name)
	}
	dest := filepath.Join(dir, t.Bin)
	if t.Release.Raw {
		return dest, os.WriteFile(dest, data, 0o755)
	}
	archive := dest + ".tar.gz"
	if err = os.WriteFile(archive, data, 0o600); err != nil {
		return "", err
	}
	return dest, extractBinary(archive, t.Bin, dest)
}
