package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type Options struct {
	Repo, Target, BinDir      string
	DryRun, JSON, Interactive bool
}
type Result struct {
	Changed bool     `json:"changed"`
	DryRun  bool     `json:"dry_run"`
	Actions []Action `json:"actions"`
	Error   string   `json:"error,omitempty"`
}
type Action struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func main() { os.Exit(run()) }
func run() int {
	f := flag.NewFlagSet("dotfiles-install", flag.ContinueOnError)
	profile := f.String("profile", "", "YAML or JSON selection (non-interactive with --yes/--dry-run)")
	tools := f.String("tools", "", "comma-separated tools; replaces profile.tools")
	cats := f.String("categories", "", "comma-separated categories; replaces profile.categories")
	exclude := f.String("exclude", "", "comma-separated exclusions; replaces profile.exclude")
	packages := f.String("packages", "", "extra native packages, comma-separated")
	yes := f.Bool("yes", false, "apply explicit selection without prompts")
	tui := f.Bool("tui", false, "edit the selection in the TUI")
	list := f.Bool("list", false, "list tools/categories without changing anything")
	dry := f.Bool("dry-run", false, "plan only; no writes, downloads or sudo")
	jsonOut := f.Bool("json", false, "machine-readable result on stdout, logs on stderr")
	noConfigs := f.Bool("no-configs", false, "install packages/binaries only")
	configsOnly := f.Bool("configs-only", false, "deploy selected configs only")
	mode := f.String("config-mode", "", "copy (default) or link")
	conflict := f.String("conflict", "", "backup (default) or error")
	repo := f.String("repo", "", "repository path (auto-detected by default)")
	home, _ := os.UserHomeDir()
	target := f.String("target", home, "target home directory; run as its owner")
	binDir := f.String("bin-dir", "", "release/compatibility binaries directory (default: TARGET/.local/bin)")
	if err := f.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	result := Result{DryRun: *dry, Actions: []Action{}}
	finish := func(err error) int {
		if err != nil {
			result.Error = err.Error()
		}
		if *jsonOut {
			_ = json.NewEncoder(os.Stdout).Encode(result)
		} else {
			for _, a := range result.Actions {
				fmt.Printf("[%s] %s: %s\n", a.Status, a.Kind, a.Name)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			} else {
				fmt.Printf("changed=%t dry_run=%t\n", result.Changed, result.DryRun)
			}
		}
		if err != nil {
			return 1
		}
		return 0
	}
	if f.NArg() != 0 {
		return finish(fmt.Errorf("unexpected positional arguments: %v", f.Args()))
	}
	c, err := loadCatalog()
	if err != nil {
		return finish(err)
	}
	if *list {
		if *jsonOut {
			returnCode := 0
			if err := json.NewEncoder(os.Stdout).Encode(c.Tools); err != nil {
				returnCode = 1
			}
			return returnCode
		}
		for _, t := range c.Tools {
			source := "package"
			if t.Release != nil {
				source = "release"
			}
			if t.External {
				source = "external/config only"
			}
			fmt.Printf("%-14s %-15s %-22s config=%-12s %s\n", t.Category, t.Name, source, t.Config, t.Description)
		}
		return 0
	}
	if *noConfigs && *configsOnly {
		return finish(fmt.Errorf("--no-configs and --configs-only are mutually exclusive"))
	}
	if *tui && (*yes || *jsonOut) {
		return finish(fmt.Errorf("--tui cannot be used with --yes or --json"))
	}
	p := defaultProfile()
	if *profile != "" {
		p, err = readProfile(*profile)
		if err != nil {
			return finish(err)
		}
	}
	f.Visit(func(v *flag.Flag) {
		switch v.Name {
		case "tools":
			p.Tools = split(*tools)
		case "categories":
			p.Categories = split(*cats)
		case "exclude":
			p.Exclude = split(*exclude)
		case "packages":
			p.Packages = split(*packages)
		}
	})
	if *noConfigs {
		p.Configs = false
	}
	if *configsOnly {
		p.Install = false
		p.Configs = true
	}
	if *mode != "" {
		p.ConfigMode = *mode
	}
	if *conflict != "" {
		p.Conflict = *conflict
	}
	selected, err := selectTools(c, p)
	if err != nil {
		return finish(err)
	}
	interactive := *tui || (!*yes && !*dry && !*jsonOut && *profile == "" && *tools == "" && *cats == "")
	if !interactive && !*yes && !*dry {
		return finish(fmt.Errorf("use --dry-run to plan, --yes to apply, or --tui to edit"))
	}
	if interactive {
		p, err = choose(c, p)
		if err != nil {
			return finish(err)
		}
		if p.Version == 0 {
			return 0
		}
		selected, err = selectTools(c, p)
		if err != nil {
			return finish(err)
		}
	}
	if len(selected) == 0 && len(p.Packages) == 0 {
		return finish(fmt.Errorf("empty selection; choose tools, categories or packages"))
	}
	if !p.Install && !p.Configs {
		return finish(fmt.Errorf("both installation and configs are disabled"))
	}
	root, err := findRepo(*repo)
	if err != nil && p.Configs {
		return finish(err)
	}
	absTarget, err := filepath.Abs(*target)
	if err != nil {
		return finish(err)
	}
	if absTarget == "/" {
		return finish(fmt.Errorf("target must be a home directory, not /"))
	}
	absBin := *binDir
	if absBin == "" {
		absBin = filepath.Join(absTarget, ".local/bin")
	}
	absBin, err = filepath.Abs(absBin)
	if err != nil {
		return finish(err)
	}
	opts := Options{Repo: root, Target: absTarget, BinDir: absBin, DryRun: *dry, JSON: *jsonOut, Interactive: interactive}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = apply(ctx, selected, p, opts, &result)
	return finish(err)
}
func split(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func findRepo(explicit string) (string, error) {
	var candidates []string
	if explicit != "" {
		candidates = []string{explicit}
	} else {
		if exe, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Dir(filepath.Dir(exe)))
		}
		cwd, _ := os.Getwd()
		candidates = append(candidates, cwd, filepath.Dir(cwd))
	}
	for _, p := range candidates {
		if s, err := os.Stat(filepath.Join(p, "config")); err == nil && s.IsDir() {
			return filepath.Abs(p)
		}
	}
	return "", fmt.Errorf("cannot find config/; pass --repo /path/to/Dotfiles")
}
