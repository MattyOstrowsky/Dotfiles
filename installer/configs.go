package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ConfigFile struct {
	Source, Dest string
	Mode         fs.FileMode
	Data         []byte
	Change       bool
}
type ConfigPlan struct {
	Files  []ConfigFile
	Detach []string
}

// Never follow a directory symlink while writing configs. Existing Stow
// directories within this repository are detached into ordinary directories;
// unrelated directory symlinks need the user to choose a different target.
func planConfigs(tools []Tool, p Profile, o Options) (ConfigPlan, error) {
	plan := ConfigPlan{}
	detach := map[string]bool{}
	seen := map[string]bool{}
	if !p.Configs {
		return plan, nil
	}
	for _, t := range tools {
		if t.Config == "" {
			continue
		}
		root := filepath.Join(o.Repo, "config", t.Config)
		err := filepath.WalkDir(root, func(src string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Name() == ".gitignore" || d.Name() == "fish_variables" || d.Name() == "navi.log" || d.Name() == ".git" || strings.HasPrefix(d.Name(), ".dotfiles-") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("config source symlinks are not supported: %s", src)
			}
			rel, err := filepath.Rel(root, src)
			if err != nil {
				return err
			}
			dest := filepath.Join(o.Target, rel)
			if seen[dest] {
				return fmt.Errorf("overlapping config destination %s", dest)
			}
			seen[dest] = true
			for dir := filepath.Dir(dest); ; dir = filepath.Dir(dir) {
				info, statErr := os.Lstat(dir)
				if statErr != nil && !os.IsNotExist(statErr) {
					return statErr
				}
				if statErr == nil {
					if info.Mode()&os.ModeSymlink != 0 {
						resolved, e := filepath.EvalSymlinks(dir)
						if e != nil {
							return e
						}
						sourceRoot, e := filepath.EvalSymlinks(root)
						if e != nil {
							return e
						}
						sourceRel, e := filepath.Rel(sourceRoot, resolved)
						if e != nil || sourceRel == ".." || strings.HasPrefix(sourceRel, ".."+string(os.PathSeparator)) {
							return fmt.Errorf("config parent %s is a symlink outside the selected config package; use a real target directory", dir)
						}
						if p.ConfigMode == "copy" {
							if p.Conflict == "error" {
								return fmt.Errorf("directory symlink conflict: %s", dir)
							}
							detach[dir] = true
						}
					} else if !info.IsDir() {
						return fmt.Errorf("config parent is not a directory: %s", dir)
					}
				}
				if dir == "/" {
					break
				}
			}
			data, e := os.ReadFile(src)
			if e != nil {
				return e
			}
			info, e := d.Info()
			if e != nil {
				return e
			}
			file := ConfigFile{Source: src, Dest: dest, Mode: info.Mode().Perm(), Data: data, Change: true}
			current, e := os.Lstat(dest)
			if e != nil && !os.IsNotExist(e) {
				return e
			}
			if e == nil {
				if current.IsDir() {
					return fmt.Errorf("config file destination is a directory: %s", dest)
				}
				if p.ConfigMode == "link" {
					resolved, linkErr := filepath.EvalSymlinks(dest)
					sourceResolved, _ := filepath.EvalSymlinks(src)
					file.Change = linkErr != nil || resolved != sourceResolved
				} else if current.Mode().IsRegular() {
					old, readErr := os.ReadFile(dest)
					if readErr != nil {
						return readErr
					}
					file.Change = !bytes.Equal(old, data) || current.Mode().Perm() != file.Mode
				}
				if file.Change && p.Conflict == "error" {
					return fmt.Errorf("config conflict: %s", dest)
				}
			}
			plan.Files = append(plan.Files, file)
			return nil
		})
		if err != nil {
			return plan, fmt.Errorf("%s config: %w", t.Name, err)
		}
	}
	for dir := range detach {
		plan.Detach = append(plan.Detach, dir)
	}
	sort.Strings(plan.Detach)
	return plan, nil
}
func backupPath(dest string) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(dest), ".dotfiles-backup-"+filepath.Base(dest)+"-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Remove(name); err != nil {
		return "", err
	}
	return name, nil
}
func applyConfigs(plan ConfigPlan, p Profile, r *Result) error {
	for _, dir := range plan.Detach {
		// Copy the previous directory, including untracked local files, before
		// replacing its symlink. os.CopyFS does not follow nested symlinks.
		source, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return err
		}
		tmp, err := os.MkdirTemp(filepath.Dir(dir), ".dotfiles-detach-*")
		if err != nil {
			return err
		}
		if err = os.CopyFS(tmp, os.DirFS(source)); err != nil {
			os.RemoveAll(tmp)
			return err
		}
		backup, err := backupPath(dir)
		if err != nil {
			os.RemoveAll(tmp)
			return err
		}
		if err = os.Rename(dir, backup); err != nil {
			os.RemoveAll(tmp)
			return err
		}
		if err = os.Rename(tmp, dir); err != nil {
			_ = os.Rename(backup, dir)
			os.RemoveAll(tmp)
			return err
		}
		r.Changed = true
		r.Actions = append(r.Actions, Action{"backup", backup, "changed"})
	}
	for _, f := range plan.Files {
		if !f.Change {
			r.Actions = append(r.Actions, Action{"config", f.Dest, "ok"})
			continue
		}
		if err := os.MkdirAll(filepath.Dir(f.Dest), 0o755); err != nil {
			return err
		}
		// Prepare the replacement on the destination filesystem before backup.
		tmp, err := os.CreateTemp(filepath.Dir(f.Dest), ".dotfiles-*")
		if err != nil {
			return err
		}
		name := tmp.Name()
		if p.ConfigMode == "copy" {
			_, err = tmp.Write(f.Data)
			if err == nil {
				err = tmp.Chmod(f.Mode)
			}
			closeErr := tmp.Close()
			if err == nil {
				err = closeErr
			}
		} else {
			err = tmp.Close()
			if err == nil {
				err = os.Remove(name)
			}
			if err == nil {
				err = os.Symlink(f.Source, name)
			}
		}
		if err != nil {
			os.Remove(name)
			return err
		}
		backup := ""
		if _, err = os.Lstat(f.Dest); err == nil {
			backup, err = backupPath(f.Dest)
			if err == nil {
				err = os.Rename(f.Dest, backup)
			}
			if err != nil {
				os.Remove(name)
				return err
			}
		} else if !os.IsNotExist(err) {
			os.Remove(name)
			return err
		}
		if err = os.Rename(name, f.Dest); err != nil {
			if backup != "" {
				_ = os.Rename(backup, f.Dest)
			}
			os.Remove(name)
			return err
		}
		r.Changed = true
		r.Actions = append(r.Actions, Action{"config", f.Dest, "changed"})
		if backup != "" {
			r.Actions = append(r.Actions, Action{"backup", backup, "changed"})
		}
	}
	return nil
}
