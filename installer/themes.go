package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// themeSpec holds, for one palette, the literal values written into each
// themable tool's repo config. applyTheme rewrites the repo files in place;
// because stow symlinks make those files the live configs, a single call
// re-renders every tool at once.
type themeSpec struct {
	Bat         string // bat config --theme="…" value
	Btop        string // btop.conf color_theme = "…"
	K9s         string // k9s config.yaml skin: …
	Nvim        string // chadrc.lua theme = "…"
	Starship    string // starship.toml palette = "…"
	Lazydocker  string // lazydocker borderColor hex (without quotes)
	Omp         string // omp config.yml theme.dark value
	Opencode    string // opencode tui.json "theme" value
	LazygitBody string // 4-space-indented body of the gui.theme sub-block (trailing newline)
}

// lazygitNordBody is the Nord gui.theme body (derived from the Nord palette;
// inactive border uses #4c566a nord-3, selection bg uses #3b4252 nord-1).
const lazygitNordBody = `    lightTheme: false
    activeBorderColor:
      - "#b48ead"
      - bold
    inactiveBorderColor:
      - "#4c566a"
    optionsTextColor:
      - "#81a1c1"
    selectedLineBgColor:
      - "#3b4252"
    selectedRangeBgColor:
      - "#3b4252"
    cherryPickedCommitBgColor:
      - "#3b4252"
      - bold
    cherryPickedCommitFgColor:
      - "#b48ead"
    unstagedChangesColor:
      - "#bf616a"
    defaultFgColor:
      - "#d8dee9"
`

// lazygitCatppuccinBody reproduces the committed Catppuccin Mocha gui.theme
// block so the default round-trips cleanly under applyTheme("catppuccin").
const lazygitCatppuccinBody = `    lightTheme: false
    activeBorderColor:
      - "#cba6f7"
      - bold
    inactiveBorderColor:
      - "#6c7086"
    optionsTextColor:
      - "#89b4fa"
    selectedLineBgColor:
      - "#313244"
    selectedRangeBgColor:
      - "#313244"
    cherryPickedCommitBgColor:
      - "#313244"
      - bold
    cherryPickedCommitFgColor:
      - "#cba6f7"
    unstagedChangesColor:
      - "#f38ba8"
    defaultFgColor:
      - "#cdd6f4"
`

var themeTable = map[string]themeSpec{
	"catppuccin": {
		Bat:         "Catppuccin Mocha",
		Btop:        "catppuccin_mocha",
		K9s:         "catppuccin",
		Nvim:        "catppuccin",
		Starship:    "catppuccin",
		Lazydocker:  "#cba6f7",
		Omp:         "dark-catppuccin",
		Opencode:    "catppuccin",
		LazygitBody: lazygitCatppuccinBody,
	},
	"nord": {
		Bat:         "Nord",
		Btop:        "nord",
		K9s:         "nord",
		Nvim:        "nord",
		Starship:    "nord",
		Lazydocker:  "#88c0d0",
		Omp:         "dark-nord",
		Opencode:    "nord",
		LazygitBody: lazygitNordBody,
	},
}

// themeDisplayNames is the ordered list for the TUI selector.
var themeDisplayNames = []string{"Catppuccin Mocha", "Nord"}

// themeKeys maps a display-name index to the themeTable key.
var themeKeys = []string{"catppuccin", "nord"}

// themeApplyResult reports per-tool outcome for human-friendly output.
type themeApplyResult struct {
	Tool   string
	Status string // "ok" (written), "skip" (not modified), "error"
	Detail string
}

// themeTarget describes one themable tool's repo config: which file under
// <repoRoot>/config/ is rewritten and the regexp that locates the palette
// value inside it, plus the literal replacement for the given palette.
type themeTarget struct {
	tool string
	path string // repo-relative path, e.g. config/bat/.config/bat/config
	re   *regexp.Regexp
	repl string
}

// themeTargets returns, in fixed tool order, every repo config a palette
// re-renders, with replacements computed from spec.
func themeTargets(spec themeSpec) []themeTarget {
	return []themeTarget{
		{"bat", "config/bat/.config/bat/config",
			regexp.MustCompile(`(?m)^--theme=".*"$`),
			fmt.Sprintf(`--theme="%s"`, spec.Bat)},
		{"btop", "config/btop/.config/btop/btop.conf",
			regexp.MustCompile(`(?m)^color_theme\s*=\s*".*"$`),
			fmt.Sprintf(`color_theme = "%s"`, spec.Btop)},
		{"k9s", "config/k9s/.config/k9s/config.yaml",
			regexp.MustCompile(`(?m)^(\s*)skin:\s.*$`),
			fmt.Sprintf(`${1}skin: %s`, spec.K9s)},
		{"nvim", "config/nvim/.config/nvim/lua/chadrc.lua",
			regexp.MustCompile(`(?m)^(\s*)theme\s*=\s*".*"`),
			fmt.Sprintf(`${1}theme = "%s"`, spec.Nvim)},
		{"starship", "config/starship/.config/starship.toml",
			regexp.MustCompile(`(?m)^palette\s*=\s*".*"$`),
			fmt.Sprintf(`palette = "%s"`, spec.Starship)},
		{"lazygit", "config/lazygit/.config/lazygit/config.yml",
			regexp.MustCompile(`(?m)^  theme:\n(?: {4}.*\n)+`),
			"  theme:\n" + spec.LazygitBody},
		{"lazydocker", "config/lazydocker/.config/lazydocker/config.yml",
			regexp.MustCompile(`(?m)^(\s*)borderColor:\s*['"].*['"]$`),
			fmt.Sprintf(`${1}borderColor: '%s'`, spec.Lazydocker)},
		{"omp", "config/omp/.omp/agent/config.yml",
			regexp.MustCompile(`(?m)^(\s*)dark:\s*(\S+)$`),
			fmt.Sprintf(`${1}dark: %s`, spec.Omp)},
		{"opencode", "config/opencode/.config/opencode/tui.json",
			regexp.MustCompile(`("theme"\s*:\s*")[^"]*(")`),
			fmt.Sprintf(`${1}%s${2}`, spec.Opencode)},
	}
}

// themePlanItem is one tool's outcome of planning a theme application. A
// missing config file is a "skip"; a present file whose content validated is
// "ready" with its fully prepared replacement in data; a read or pattern
// failure is an "error". applyTheme reuses the same items for execution,
// turning "ready" into "ok" (written), "error" (write failed) or "skip" (a
// later failure aborted the run before this write).
type themePlanItem struct {
	tool   string
	rel    string // repo-relative path, for messages
	full   string // absolute path, for read/write
	status string // "ready", "skip", "ok", "error"
	detail string // human detail for skip/error rows
	errMsg string // one aggregate-error line; set when status is "error"
	data   []byte // prepared replacement content; nil unless ready/ok
}

// planThemeApply validates that the named theme can be applied to every
// present repo config and prepares all rewrites in memory, writing nothing.
// Each tool's file is read and matched against its expected shape; missing
// files are planned as skips, read/pattern failures as errors. A plan with any
// error must not be executed, since nothing has been touched yet. Future UI
// code can call it to preview whether a theme is applicable (and which tools
// would be skipped) before mutating the repo.
func planThemeApply(name, repoRoot string) ([]themePlanItem, error) {
	spec, ok := themeTable[name]
	if !ok {
		return nil, fmt.Errorf("unknown theme %q (want one of: catppuccin, nord)", name)
	}
	targets := themeTargets(spec)
	plan := make([]themePlanItem, 0, len(targets))
	for _, t := range targets {
		item := themePlanItem{tool: t.tool, rel: t.path, full: filepath.Join(repoRoot, t.path)}
		data, err := os.ReadFile(item.full)
		switch {
		case os.IsNotExist(err):
			item.status = "skip"
			item.detail = "file missing"
		case err != nil:
			item.status = "error"
			item.detail = err.Error()
			item.errMsg = fmt.Sprintf("%s: read: %v", item.tool, err)
		case !t.re.Match(data):
			item.status = "error"
			item.detail = fmt.Sprintf("pattern not found in %s (config shape changed?)", item.rel)
			item.errMsg = fmt.Sprintf("%s: %s", item.tool, item.detail)
		default:
			item.status = "ready"
			item.data = t.re.ReplaceAll(data, []byte(t.repl))
		}
		plan = append(plan, item)
	}
	return plan, nil
}

// applyTheme rewrites every themable tool's repo config (under repoRoot) to
// the named palette. All present files are first read, validated against their
// expected shape and re-rendered in memory (see planThemeApply); an unknown
// theme, an unreadable file or a zero-match aborts the call before anything is
// written, so the repo is never partially re-themed on validation failure.
// Missing files are skipped (recorded, non-fatal). Only after every present
// file validated are the prepared contents written, in tool order; a write
// failure stops the run so no subsequent file is changed. Returns per-tool
// results for reporting plus an aggregate error listing any failures.
func applyTheme(name, repoRoot string) ([]themeApplyResult, error) {
	plan, err := planThemeApply(name, repoRoot)
	if err != nil {
		return nil, err
	}

	// Validation must pass for every tool before the first write.
	abort := false
	for i := range plan {
		if plan[i].status == "error" {
			abort = true
			break
		}
	}
	if abort {
		for i := range plan {
			if plan[i].status == "ready" {
				plan[i].status = "skip"
				plan[i].detail = "not applied (validation failed before any write)"
			}
		}
	} else {
		// All files validated: write the prepared contents in tool order. A
		// write failure aborts the remaining writes so no subsequent file
		// changes; later ready tools are reported as skips.
		abortedBy := ""
		for i := range plan {
			item := &plan[i]
			if item.status != "ready" { // missing files stay reported as skips
				continue
			}
			if abortedBy != "" {
				item.status = "skip"
				item.detail = fmt.Sprintf("not applied (aborted after %s write error)", abortedBy)
				continue
			}
			if err := os.WriteFile(item.full, item.data, 0o644); err != nil {
				item.status = "error"
				item.detail = err.Error()
				item.errMsg = fmt.Sprintf("%s: write: %v", item.tool, err)
				abortedBy = item.tool
				continue
			}
			item.status = "ok"
		}
	}

	results := make([]themeApplyResult, 0, len(plan))
	errs := make([]string, 0, len(plan))
	for _, item := range plan {
		results = append(results, themeApplyResult{item.tool, item.status, item.detail})
		if item.status == "error" {
			errs = append(errs, item.errMsg)
		}
	}
	agg := error(nil)
	if len(errs) > 0 {
		agg = fmt.Errorf("%d tool(s) failed:\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return results, agg
}
