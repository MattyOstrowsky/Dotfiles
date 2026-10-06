package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractBinaryIgnoresArchivePathsAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "tool.tgz")
	f, _ := os.Create(archive)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, h := range []*tar.Header{{Name: "tool", Typeflag: tar.TypeSymlink, Linkname: "/tmp/unsafe"}, {Name: "../../tool", Mode: 0o755, Size: 4, Typeflag: tar.TypeReg}} {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			tw.Write([]byte("test"))
		}
	}
	tw.Close()
	gz.Close()
	f.Close()
	dest := filepath.Join(dir, "binary")
	if err := extractBinary(archive, "tool", dest); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "test" {
		t.Fatal(string(data))
	}
	if err := extractBinary(archive, "missing", filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing binary accepted")
	}
}
