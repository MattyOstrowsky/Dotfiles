package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	listFlag := flag.Bool("list", false, "print detected system and tool/config status without changing anything")
	yesFlag := flag.Bool("yes", false, "non-interactive: install all tools and applicable config packages")
	preflightFlag := flag.Bool("preflight", false, "refresh package metadata and print the installation plan")
	validateFlag := flag.Bool("validate", false, "scan installed tools and config links without changing anything")
	themeFlag := flag.String("theme", "", "apply a theme (catppuccin|nord); with --yes apply it during installation")
	flag.Parse()

	if *themeFlag != "" {
		if _, ok := themeTable[*themeFlag]; !ok {
			fatal(fmt.Errorf("unknown theme %q (want one of: catppuccin, nord)", *themeFlag))
		}
	}
	if *listFlag && (*yesFlag || *preflightFlag || *validateFlag) {
		fatal(fmt.Errorf("--list cannot be combined with --yes, --preflight or --validate"))
	}
	if *preflightFlag && (*yesFlag || *validateFlag) {
		fatal(fmt.Errorf("--preflight cannot be combined with --yes or --validate"))
	}
	if *validateFlag && *yesFlag {
		fatal(fmt.Errorf("--validate cannot be combined with --yes"))
	}

	m, err := loadManifest()
	if err != nil {
		fatal(err)
	}
	repoRoot, err := findRepoRoot()
	if err != nil {
		fatal(err)
	}
	pkgs, err := scanStowPkgs(repoRoot)
	if err != nil {
		fatal(err)
	}

	switch {
	case *listFlag:
		os.Exit(runList(m, pkgs, repoRoot))
	case *preflightFlag:
		os.Exit(runPreflightCLI(m, pkgs, repoRoot))
	case *validateFlag:
		os.Exit(runValidateCLI(m, pkgs, repoRoot))
	case *themeFlag != "" && *yesFlag:
		os.Exit(runHeadlessTheme(m, pkgs, repoRoot, *themeFlag))
	case *themeFlag != "":
		os.Exit(runThemeOnly(*themeFlag, repoRoot))
	case *yesFlag:
		os.Exit(runHeadless(m, pkgs, repoRoot))
	default:
		os.Exit(runTUI(m, pkgs, repoRoot))
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func printTables(m *Manifest, pkgs []StowPkg) {
	fmt.Println("TOOLS")
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  name\tsource\tdeps\tstow pkg")
	for _, t := range m.Tools {
		stow := t.Stow
		if stow == "" {
			stow = "—"
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", t.Name, t.SourceLabel(), t.DepsLabel(), stow)
	}
	w.Flush()

	fmt.Println("\nSTOW PACKAGES (config/)")
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  package\ttarget(s)")
	for _, p := range pkgs {
		fmt.Fprintf(w, "  %s\t%s\n", p.Name, strings.Join(p.Mappings, ", "))
	}
	w.Flush()
}

func installTarget() (string, error) {
	target, err := os.UserHomeDir()
	if err != nil || target == "" {
		if err == nil {
			err = fmt.Errorf("home directory is empty")
		}
		return "", fmt.Errorf("resolve install target: %w", err)
	}
	return target, nil
}

// applicableConfigPkgs excludes Windows-side WSL packages when installing into
// the Linux home. A caller targeting a separate Windows tree may retain them.
func applicableConfigPkgs(pkgs []StowPkg, target string) []StowPkg {
	target = expandTarget(target)
	out := make([]StowPkg, 0, len(pkgs))
	for _, p := range pkgs {
		host := p.Host
		if host == "" {
			host = pkgHostScope(p.Name)
		}
		if configTargetSupported(host, target) {
			p.Host = host
			out = append(out, p)
		}
	}
	return out
}

func printSystem(system SystemContext) {
	fmt.Printf("SYSTEM\n  distro: %s\n  manager: %s\n  arch: %s\n", system.Distro, system.Kind(), system.Arch)
}

func printPreflightReport(report PreflightReport) {
	printSystem(report.System)
	fmt.Println("\nTOOLS")
	for _, result := range report.ToolResults {
		fmt.Printf("  [%s] %s", result.Status, result.Tool.Name)
		if result.Detail != "" {
			fmt.Printf(" — %s", result.Detail)
		}
		fmt.Println()
		for _, pkg := range result.Packages {
			fmt.Printf("      %s: [%s] %s\n", pkg.Name, pkg.Status, pkg.Detail)
		}
	}
	fmt.Println("\nCONFIGS")
	for _, result := range report.ConfigPlans {
		fmt.Printf("  [%s] %s", result.Status, result.Config.Name)
		if result.Detail != "" {
			fmt.Printf(" — %s", result.Detail)
		}
		fmt.Println()
	}
	for _, warning := range report.Warnings {
		fmt.Printf("  ! %s\n", warning)
	}
	for _, err := range report.Errors {
		fmt.Fprintf(os.Stderr, "  ✗ %v\n", err)
	}
}

func runList(m *Manifest, pkgs []StowPkg, repoRoot string) int {
	system, err := detectSystem()
	if err != nil {
		fmt.Fprintln(os.Stderr, "system detection:", err)
		return 1
	}
	target, err := installTarget()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	pkgs = applicableConfigPkgs(pkgs, target)
	report := BuildPreflight(system, m.Tools, pkgs, repoRoot, target, false)
	printPreflightReport(report)
	return 0
}

func runPreflightCLI(m *Manifest, pkgs []StowPkg, repoRoot string) int {
	system, err := detectSystem()
	if err != nil {
		fmt.Fprintln(os.Stderr, "system detection:", err)
		return 1
	}
	target, err := installTarget()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := prepareSudo(true); err != nil {
		fmt.Fprintln(os.Stderr, "preflight:", err)
		return 1
	}
	report := BuildPreflight(system, m.Tools, applicableConfigPkgs(pkgs, target), repoRoot, target, true)
	printPreflightReport(report)
	if len(report.Errors) > 0 {
		return 1
	}
	return 0
}

func runValidateCLI(m *Manifest, pkgs []StowPkg, repoRoot string) int {
	system, err := detectSystem()
	if err != nil {
		fmt.Fprintln(os.Stderr, "system detection:", err)
		return 1
	}
	target, err := installTarget()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	report := ValidateInstallation(system, m.Tools, applicableConfigPkgs(pkgs, target), target)
	printSystem(system)
	for _, tool := range report.Tools {
		fmt.Printf("  [%s] %s — %s\n", tool.Status, tool.Tool, tool.Detail)
	}
	for _, config := range report.Configs {
		fmt.Printf("  [%s] config %s — %s\n", config.Status, config.Config.Name, config.Detail)
	}
	for _, warning := range report.Warnings {
		fmt.Printf("  ! %s\n", warning)
	}
	for _, err := range report.Errors {
		fmt.Fprintf(os.Stderr, "  ✗ %v\n", err)
	}
	if !report.Valid || len(report.Errors) > 0 {
		return 1
	}
	return 0
}

// osCommandSudoV validates sudo credentials interactively (password prompt).
func osCommandSudoV() ([]byte, error) {
	return exec.Command("sudo", "-v").CombinedOutput()
}

// sudoReady reports whether privileged commands can run without a password
// prompt: cached credentials (sudo -n -v) or NOPASSWD rules (sudo -n true,
// which also covers sudo-rs where -v ignores NOPASSWD command entries).
func sudoReady() bool {
	if err := exec.Command("sudo", "-n", "-v").Run(); err == nil {
		return true
	}
	return exec.Command("sudo", "-n", "true").Run() == nil
}

// keepSudoWarm refreshes sudo credentials every minute so long install
// phases never hit a password prompt inside the TUI.
func keepSudoWarm() {
	for range time.Tick(60 * time.Second) {
		exec.Command("sudo", "-n", "-v").Run()
	}
}

func prepareSudo(required bool) error {
	if !required || sudoReady() {
		if required {
			go keepSudoWarm()
		}
		return nil
	}
	if out, err := osCommandSudoV(); err != nil {
		return fmt.Errorf("sudo -v failed: %w\n%s", err, tail(string(out), 5))
	}
	go keepSudoWarm()
	return nil
}

// preflight refreshes the detected native package manager and ensures the
// local binary directory. It is retained for the TUI's early terminal phase.
func preflight(log func(string)) error {
	say := func(s string) {
		if log != nil {
			log(s)
		}
	}
	system, err := detectSystem()
	if err != nil {
		return err
	}
	say("sudo check …")
	if err := prepareSudo(true); err != nil {
		return err
	}
	say(string(system.Kind()) + " metadata refresh …")
	if err := system.Manager.Refresh(); err != nil {
		return err
	}
	say("mkdir -p ~/.local/bin …")
	if err := os.MkdirAll(localBin(), 0o755); err != nil {
		return err
	}
	say("preflight OK")
	return nil
}

// runHeadless implements --yes: both phases, everything selected.
// runHeadless implements --yes: planner, executor, stow and post-validation.
func runHeadless(m *Manifest, pkgs []StowPkg, repoRoot string) int {
	return runHeadlessWithTheme(m, pkgs, repoRoot, "catppuccin")
}

func runHeadlessWithTheme(m *Manifest, pkgs []StowPkg, repoRoot, theme string) int {
	system, err := detectSystem()
	if err != nil {
		fmt.Fprintln(os.Stderr, "system detection:", err)
		return 1
	}
	target, err := installTarget()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	tools := append([]Tool(nil), m.Tools...)
	configPkgs := applicableConfigPkgs(pkgs, target)
	needsNative := len(NativePackageUnion(tools, system.Kind())) > 0
	if err := prepareSudo(needsNative); err != nil {
		fmt.Fprintln(os.Stderr, "sudo:", err)
		return 1
	}
	if err := os.MkdirAll(localBin(), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "local bin:", err)
		return 1
	}

	fmt.Println("== Preflight ==")
	plan := BuildPreflight(system, tools, configPkgs, repoRoot, target, needsNative)
	printPreflightReport(plan)
	if len(plan.Errors) > 0 {
		return 1
	}

	fmt.Println("\n== Tools phase ==")
	execReport := ExecuteTools(context.Background(), system, tools, func(event ExecutionEvent) {
		switch event.Status {
		case ExecutionRunning:
			fmt.Printf("  … %s", event.Name)
		case ExecutionOK:
			fmt.Printf("  ✓ %s", event.Name)
		case ExecutionFail:
			fmt.Printf("  ✗ %s", event.Name)
		case ExecutionSkip:
			fmt.Printf("  − %s", event.Name)
		default:
			return
		}
		if event.Detail != "" {
			fmt.Printf(" — %s", event.Detail)
		}
		fmt.Println()
	})
	for _, execErr := range execReport.Errors {
		fmt.Fprintln(os.Stderr, "tools:", execErr)
	}
	failures := execReport.Failed

	for i := range tools {
		if tools[i].Name == "fish" && tools[i].Installed() && !strings.Contains(os.Getenv("SHELL"), "fish") {
			if err := chshFish(); err != nil {
				fmt.Println("  ! chsh to fish failed (fish still works via `fish`):", err)
			} else {
				fmt.Println("  ✓ login shell set to fish")
			}
		}
	}

	fmt.Println("\n== Configs phase ==")
	for i := range configPkgs {
		p := &configPkgs[i]
		backup, applyErr := applyStow(repoRoot, p, target)
		if applyErr != nil {
			fmt.Printf("  ✗ %s: %v\n", p.Name, applyErr)
			failures = true
			continue
		}
		if backup != "" {
			fmt.Printf("  ✓ %s (conflicts backed up to %s)\n", p.Name, backup)
		} else {
			fmt.Printf("  ✓ %s\n", p.Name)
		}
	}

	if !failures {
		fmt.Println("\n== Theme ==")
		if code := printThemeApply(theme, repoRoot); code != 0 {
			failures = true
		}
	}

	fmt.Println("\n== Validation ==")
	validation := ValidatePostInstall(system, tools, configPkgs, target)
	for _, tool := range validation.Tools {
		fmt.Printf("  [%s] %s — %s\n", tool.Status, tool.Tool, tool.Detail)
	}
	for _, config := range validation.Configs {
		fmt.Printf("  [%s] config %s — %s\n", config.Status, config.Config.Name, config.Detail)
	}
	if !validation.Valid || len(validation.Errors) > 0 {
		failures = true
	}
	if failures {
		fmt.Println("\nFinished with failures (see above).")
		return 1
	}
	fmt.Println("\nAll done.")
	return 0
}

// runHeadlessTheme performs one installation and applies only the requested
// theme; it deliberately does not land on Catppuccin first.
func runHeadlessTheme(m *Manifest, pkgs []StowPkg, repoRoot, theme string) int {
	return runHeadlessWithTheme(m, pkgs, repoRoot, theme)
}

// runThemeOnly applies a theme without preflight/tools/stow — the repo
// configs are edited in place; if already stowed the symlinks propagate the
// change live.
func runThemeOnly(theme, repoRoot string) int {
	fmt.Println("== Theme ==")
	return printThemeApply(theme, repoRoot)
}

// printThemeApply runs applyTheme and prints a per-tool report. Returns 0 on
// success, 1 if any tool errored (missing files are reported but non-fatal).
func printThemeApply(theme, repoRoot string) int {
	results, err := applyTheme(theme, repoRoot)
	for _, r := range results {
		switch r.Status {
		case "ok":
			fmt.Printf("  ✓ %s\n", r.Tool)
		case "skip":
			fmt.Printf("  − %s  %s\n", r.Tool, r.Detail)
		default:
			fmt.Printf("  ✗ %s  %s\n", r.Tool, r.Detail)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "theme apply:", err)
		return 1
	}
	fmt.Printf("  theme: %s\n", theme)
	return 0
}
