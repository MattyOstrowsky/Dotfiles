package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// screen is intentionally small: all interactive work happens in the dashboard,
// while the target editor and result page remain explicit states.
type screen int

const (
	scrPreflight screen = iota
	scrDashboard
	scrConfigTarget
	scrDone
)

type dashboardTab int

const (
	tabTools dashboardTab = iota
	tabConfigs
	tabValidate
	tabTheme
	tabSummary
)

var tabNames = []string{"Tools", "Configs", "Validate", "Theme", "Summary"}

// Messages contain only data produced by a tea.Cmd. Commands never retain a
// pointer to model and therefore cannot race with Update or View.
type preflightDoneMsg struct {
	report PreflightReport
	err    error
}
type toolsDoneMsg struct{ report ExecutionReport }
type configsDoneMsg struct {
	results []configApplyResult
	target  string
}
type validateDoneMsg struct{ report ValidationReport }
type themeDoneMsg struct {
	name    string
	results []themeApplyResult
	err     error
}

type configApplyResult struct {
	index  int
	status stepStatus
	detail string
}

type stepStatus int

const (
	stPending stepStatus = iota
	stRunning
	stOK
	stFail
	stSkip
)

var (
	styTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#5f4b8b", Dark: "#cba6f7"})
	styAccent = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0969a5", Dark: "#89b4fa"})
	styOK     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#18753c", Dark: "#a6e3a1"})
	styFail   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#b42318", Dark: "#f38ba8"})
	styWarn   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8a5700", Dark: "#f9e2af"})
	styRun    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9a6700", Dark: "#f2cd7d"})
	styDim    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#667085", Dark: "#6c7086"})
	styBold   = lipgloss.NewStyle().Bold(true)
	styRule   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#98a2b3", Dark: "#585b70"})
)

type model struct {
	manifest *Manifest
	pkgs     []StowPkg
	repoRoot string
	system   SystemContext
	mode     string // "tools" | "configs" | "both" | "theme"

	screen screen
	tab    dashboardTab
	cursor int
	width  int
	height int
	spin   spinner.Model

	target      string
	targetInput textinput.Model
	working     bool
	workLabel   string
	started     bool

	preflight    PreflightReport
	preflightErr error
	validation   *ValidationReport
	themeName    string
	themeResults []themeApplyResult
	themeErr     error
	failures     bool

	toolSelected []bool
	toolStatus   []stepStatus
	toolDetail   []string
	pkgSelected  []bool
	pkgStatus    []stepStatus
	pkgDetail    []string

	log    viewport.Model
	detail viewport.Model
}

func newModel(manifest *Manifest, pkgs []StowPkg, repoRoot string) *model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	system, detectErr := detectSystem()
	target, _ := os.UserHomeDir()
	if target == "" {
		target = "."
	}
	log := viewport.New(70, 8)
	detail := viewport.New(70, 8)
	n := len(manifest.Tools)
	m := &model{
		manifest: manifest, pkgs: pkgs, repoRoot: repoRoot, system: system,
		mode: "both", screen: scrPreflight, target: target, spin: sp,
		width: 110, height: 32, toolSelected: make([]bool, n),
		toolStatus: make([]stepStatus, n), toolDetail: make([]string, n),
		pkgSelected: make([]bool, len(pkgs)), pkgStatus: make([]stepStatus, len(pkgs)),
		pkgDetail: make([]string, len(pkgs)), log: log, detail: detail,
	}
	if detectErr != nil {
		m.preflightErr = detectErr
	}
	m.setViewportSizes()
	return m
}

func (m *model) Init() tea.Cmd {
	// The initial plan is deliberately read-only. Refresh is an explicit user
	// action (r), never an unconditional apt/dnf/pacman update.
	return tea.Batch(m.spin.Tick, m.preflightCmd(false))
}

func (m *model) preflightCmd(refresh bool) tea.Cmd {
	ctx := m.system
	tools := append([]Tool(nil), m.manifest.Tools...)
	pkgs := append([]StowPkg(nil), m.pkgs...)
	repoRoot, target := m.repoRoot, m.target
	return func() tea.Msg {
		report := BuildPreflight(ctx, tools, pkgs, repoRoot, target, refresh)
		return preflightDoneMsg{report: report}
	}
}

func (m *model) toolsCmd() tea.Cmd {
	selected := make([]Tool, 0, len(m.manifest.Tools))
	for i, yes := range m.toolSelected {
		if yes {
			selected = append(selected, m.manifest.Tools[i])
		}
	}
	system := m.system
	return func() tea.Msg {
		report := ExecuteTools(context.Background(), system, selected, nil)
		return toolsDoneMsg{report: report}
	}
}

func (m *model) configsCmd() tea.Cmd {
	selected := make([]StowPkg, 0, len(m.pkgs))
	indices := make([]int, 0, len(m.pkgs))
	for i, yes := range m.pkgSelected {
		if yes {
			selected = append(selected, m.pkgs[i])
			indices = append(indices, i)
		}
	}
	repoRoot, target := m.repoRoot, m.target
	return func() tea.Msg {
		results := make([]configApplyResult, 0, len(selected))
		for i, pkg := range selected {
			result := configApplyResult{index: indices[i], status: stRunning}
			backup, err := applyStow(repoRoot, &pkg, target)
			if err != nil {
				result.status, result.detail = stFail, tail(err.Error(), 3)
			} else {
				result.status = stOK
				result.detail = strings.Join(pkg.Mappings, ", ")
				if backup != "" {
					result.detail = "backed up conflicts → " + backup
				}
			}
			results = append(results, result)
		}
		return configsDoneMsg{results: results, target: target}
	}
}

func (m *model) validateCmd() tea.Cmd {
	target := m.target
	tools := make([]Tool, 0, len(m.manifest.Tools))
	pkgs := make([]StowPkg, 0, len(m.pkgs))
	for i, yes := range m.toolSelected {
		if yes {
			tools = append(tools, m.manifest.Tools[i])
		}
	}
	for i, yes := range m.pkgSelected {
		if yes {
			pkgs = append(pkgs, m.pkgs[i])
		}
	}
	system := m.system
	configOnly := m.mode == "configs"
	return func() tea.Msg {
		var report ValidationReport
		if configOnly {
			report = ValidateConfigs(pkgs, target)
		} else {
			report = ValidateInstallation(system, tools, pkgs, target)
		}
		return validateDoneMsg{report: report}
	}
}

func (m *model) themeCmd(name string) tea.Cmd {
	repoRoot := m.repoRoot
	return func() tea.Msg {
		results, err := applyTheme(name, repoRoot)
		return themeDoneMsg{name: name, results: results, err: err}
	}
}

func (m *model) setViewportSizes() {
	w := m.width - 8
	if w < 24 {
		w = 24
	}
	h := m.height - 12
	if h < 3 {
		h = 3
	}
	// Keep both panes independently scrollable: detail shows the selected
	// diff/plan and log shows warnings and post-validation.
	m.log.Width, m.log.Height = w, h/2
	if m.log.Height < 3 {
		m.log.Height = 3
	}
	m.detail.Width, m.detail.Height = w, h-m.log.Height-1
	if m.detail.Height < 3 {
		m.detail.Height = 3
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case preflightDoneMsg:
		m.preflight = msg.report
		if msg.err != nil {
			m.preflightErr = msg.err
			m.failures = true
		} else if len(msg.report.Errors) > 0 {
			m.preflightErr = msg.report.Errors[0]
			m.failures = true
		}
		m.applyPreflight()
		m.screen, m.working, m.started = scrDashboard, false, true
		m.refreshViews()
		return m, nil
	case toolsDoneMsg:
		m.working = false
		m.applyToolReport(msg.report)
		if msg.report.Failed {
			m.failures = true
		}
		m.workLabel = ""
		// Post-validation is automatic after a mutating phase, so Summary always
		// contains a concrete result rather than asking the user to infer it.
		m.working, m.workLabel = true, "Validating"
		return m, m.validateCmd()
	case configsDoneMsg:
		m.working = false
		m.target = msg.target
		for _, result := range msg.results {
			m.pkgStatus[result.index], m.pkgDetail[result.index] = result.status, result.detail
			if result.status == stFail {
				m.failures = true
			}
		}
		m.workLabel = ""
		m.working, m.workLabel = true, "Validating"
		return m, m.validateCmd()
	case validateDoneMsg:
		m.working = false
		m.validation = &msg.report
		if !msg.report.Valid {
			m.failures = true
		}
		m.workLabel = ""
		m.tab = tabSummary
		m.refreshViews()
		return m, nil
	case themeDoneMsg:
		m.working = false
		m.themeName, m.themeResults, m.themeErr = msg.name, msg.results, msg.err
		if msg.err != nil {
			m.failures = true
		}
		m.tab = tabSummary
		m.workLabel = ""
		m.refreshViews()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" && !m.working && m.screen != scrConfigTarget {
			return m, tea.Quit
		}
		if m.working {
			// Navigation and selection are frozen while a Cmd is executing.
			return m, nil
		}
		return m.handleKey(msg)
	}
	if m.screen == scrConfigTarget {
		var cmd tea.Cmd
		m.targetInput, cmd = m.targetInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) applyPreflight() {
	m.toolSelected = make([]bool, len(m.manifest.Tools))
	m.toolStatus = make([]stepStatus, len(m.manifest.Tools))
	m.toolDetail = make([]string, len(m.manifest.Tools))
	for i, result := range m.preflight.ToolResults {
		m.toolSelected[i] = result.Status != PreflightInstalled && result.Status != PreflightUnavailable && result.Status != PreflightUnsupported
		m.toolDetail[i] = result.Detail
		if result.Status == PreflightInstalled {
			m.toolStatus[i] = stOK
		} else if result.Status == PreflightUnavailable || result.Status == PreflightUnsupported {
			m.toolStatus[i] = stSkip
		}
	}
	m.pkgSelected = make([]bool, len(m.pkgs))
	m.pkgStatus = make([]stepStatus, len(m.pkgs))
	m.pkgDetail = make([]string, len(m.pkgs))
	for i, result := range m.preflight.ConfigPlans {
		// HostWSL is intentionally not selected for a Linux HOME target.
		m.pkgSelected[i] = !result.Skipped && result.Config.Host != HostWSL
		m.pkgDetail[i] = result.Detail
		if result.Skipped {
			m.pkgStatus[i] = stSkip
		}
	}
	m.target = m.preflight.Target
}

func (m *model) applyToolReport(report ExecutionReport) {
	for _, event := range report.Events {
		// ExecuteTools receives only selected tools; map by stable name rather
		// than assuming selected indexes are contiguous.
		for i := range m.manifest.Tools {
			if m.manifest.Tools[i].Name != event.Name {
				continue
			}
			switch event.Status {
			case ExecutionPending:
				m.toolStatus[i] = stPending
			case ExecutionRunning:
				m.toolStatus[i] = stRunning
			case ExecutionOK:
				m.toolStatus[i] = stOK
			case ExecutionFail:
				m.toolStatus[i] = stFail
			case ExecutionSkip:
				m.toolStatus[i] = stSkip
			}
			if event.Detail != "" {
				m.toolDetail[i] = event.Detail
			}
		}
	}
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.screen == scrConfigTarget {
		switch key {
		case "enter":
			target := expandTarget(strings.TrimSpace(m.targetInput.Value()))
			if target == "" {
				target = m.target
			}
			if err := os.MkdirAll(target, 0o755); err != nil {
				m.preflightErr = err
				return m, nil
			}
			// Re-plan against exactly the chosen target before applying stow.
			m.target = target
			m.preflight = BuildPreflight(m.system, m.manifest.Tools, m.pkgs, m.repoRoot, target, false)
			m.applyPreflightTargetOnly()
			m.screen = scrDashboard
			m.tab, m.cursor = tabConfigs, 0
			return m, nil
		case "esc":
			m.screen = scrDashboard
			return m, nil
		}
		var cmd tea.Cmd
		m.targetInput, cmd = m.targetInput.Update(msg)
		return m, cmd
	}
	if m.screen == scrDone {
		if key == "enter" || key == "esc" {
			m.screen = scrDashboard
			return m, nil
		}
	}
	if key == "r" {
		m.working, m.workLabel = true, "Refreshing package metadata"
		return m, m.preflightCmd(true)
	}
	if key == "tab" || key == "right" || key == "l" {
		m.nextTab(1)
		return m, nil
	}
	if key == "shift+tab" || key == "left" || key == "h" {
		m.nextTab(-1)
		return m, nil
	}
	rows := m.rowCount()
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < rows-1 {
			m.cursor++
		}
	case " ":
		m.toggleCurrent()
	case "a":
		m.toggleAll()
	case "enter":
		return m.activateTab()
	case "esc":
		m.tab = tabSummary
		m.cursor = 0
	}
	m.refreshViews()
	return m, nil
}

func (m *model) tabAllowed(t dashboardTab) bool {
	switch m.mode {
	case "tools":
		return t == tabTools || t == tabValidate || t == tabTheme || t == tabSummary
	case "configs":
		return t == tabConfigs || t == tabValidate || t == tabTheme || t == tabSummary
	case "theme":
		return t == tabTheme || t == tabSummary
	default:
		return true
	}
}

func (m *model) nextTab(delta int) {
	for n := 0; n < len(tabNames); n++ {
		m.tab = dashboardTab((int(m.tab) + delta + len(tabNames)) % len(tabNames))
		if m.tabAllowed(m.tab) {
			m.cursor = 0
			return
		}
	}
}

func (m *model) rowCount() int {
	switch m.tab {
	case tabTools:
		return len(m.manifest.Tools)
	case tabConfigs:
		return len(m.pkgs)
	case tabTheme:
		return len(themeDisplayNames)
	default:
		return 1
	}
}

func (m *model) toggleCurrent() {
	switch m.tab {
	case tabTools:
		if m.cursor < len(m.toolSelected) && m.manifest.Tools[m.cursor].Name != "stow" {
			m.toolSelected[m.cursor] = !m.toolSelected[m.cursor]
		}
	case tabConfigs:
		if m.cursor < len(m.pkgSelected) && m.pkgStateAllowed(m.cursor) {
			m.pkgSelected[m.cursor] = !m.pkgSelected[m.cursor]
		}
	}
}

func (m *model) pkgStateAllowed(i int) bool {
	return i >= 0 && i < len(m.preflight.ConfigPlans) && !m.preflight.ConfigPlans[i].Skipped
}

func (m *model) toggleAll() {
	switch m.tab {
	case tabTools:
		all := true
		for i, yes := range m.toolSelected {
			if m.manifest.Tools[i].Name != "stow" && !yes {
				all = false
			}
		}
		for i := range m.toolSelected {
			if m.manifest.Tools[i].Name != "stow" {
				m.toolSelected[i] = !all
			}
		}
	case tabConfigs:
		all := true
		for i := range m.pkgSelected {
			if m.pkgStateAllowed(i) && !m.pkgSelected[i] {
				all = false
			}
		}
		for i := range m.pkgSelected {
			if m.pkgStateAllowed(i) {
				m.pkgSelected[i] = !all
			}
		}
	}
}

func (m *model) activateTab() (tea.Model, tea.Cmd) {
	switch m.tab {
	case tabTools:
		m.working, m.workLabel = true, "Installing tools"
		return m, m.toolsCmd()
	case tabConfigs:
		m.openTarget()
		return m, nil
	case tabValidate:
		m.working, m.workLabel = true, "Validating"
		return m, m.validateCmd()
	case tabTheme:
		if m.cursor >= len(themeKeys) {
			return m, nil
		}
		m.working, m.workLabel = true, "Applying theme"
		return m, m.themeCmd(themeKeys[m.cursor])
	case tabSummary:
		m.screen = scrDone
	}
	return m, nil
}

func (m *model) openTarget() {
	ti := textinput.New()
	ti.Prompt = "Target: "
	ti.SetValue(m.target)
	ti.CharLimit = 256
	ti.Width = m.width - 18
	if ti.Width < 20 {
		ti.Width = 20
	}
	ti.Focus()
	m.targetInput = ti
	m.screen = scrConfigTarget
}

func (m *model) applyPreflightTargetOnly() {
	for i, result := range m.preflight.ConfigPlans {
		m.pkgSelected[i] = !result.Skipped && result.Config.Host != HostWSL
		m.pkgDetail[i] = result.Detail
		m.pkgStatus[i] = stPending
		if result.Skipped {
			m.pkgStatus[i] = stSkip
		}
	}
}

func (m *model) refreshViews() {
	m.log.SetContent(m.logText())
	m.detail.SetContent(m.detailText())
}

func (m *model) logText() string {
	var b strings.Builder
	if m.preflightErr != nil {
		b.WriteString(styFail.Render("system detection: "+m.preflightErr.Error()) + "\n")
	}
	for _, warning := range m.preflight.Warnings {
		b.WriteString(styWarn.Render("! "+warning) + "\n")
	}
	if m.validation != nil {
		for _, warning := range m.validation.Warnings {
			b.WriteString(styWarn.Render("! "+warning) + "\n")
		}
	}
	if len(m.themeResults) > 0 {
		for _, result := range m.themeResults {
			b.WriteString(result.Tool + ": " + result.Status + " " + result.Detail + "\n")
		}
	}
	if b.Len() == 0 {
		b.WriteString(styDim.Render("No warnings. Press r to refresh package metadata.") + "\n")
	}
	return b.String()
}

func (m *model) detailText() string {
	var b strings.Builder
	switch m.tab {
	case tabTools:
		if m.cursor < len(m.preflight.ToolResults) {
			r := m.preflight.ToolResults[m.cursor]
			b.WriteString(styBold.Render(r.Tool.Name) + "\n\n")
			b.WriteString("source: " + r.Tool.SourceLabelFor(m.system.Kind()) + "\n")
			b.WriteString("status: " + statusLabel(r.Status) + "\n")
			if r.Detail != "" {
				b.WriteString(r.Detail + "\n")
			}
			b.WriteString("\nPackage plan:\n")
			for _, p := range r.Packages {
				b.WriteString("  " + p.Name + "  " + statusLabel(p.Status) + "  " + p.Detail + "\n")
			}
		}
	case tabConfigs:
		if m.cursor < len(m.preflight.ConfigPlans) {
			r := m.preflight.ConfigPlans[m.cursor]
			b.WriteString(styBold.Render(r.Config.Name) + "  " + string(r.Config.Host) + "\n")
			if r.Config.Desc != "" {
				b.WriteString(r.Config.Desc + "\n")
			}
			b.WriteString("target: " + m.target + "\nstatus: " + statusLabel(r.Status) + "\n")
			if r.Detail != "" {
				b.WriteString(r.Detail + "\n")
			}
			if r.Stow != nil {
				b.WriteString("\nFiles / diff:\n")
				for _, file := range r.Stow.Files {
					line := "  " + string(file.Action) + "  " + file.Rel
					if file.Detail != "" {
						line += " (" + file.Detail + ")"
					}
					b.WriteString(line + "\n")
				}
				if len(r.Stow.BackupRels) > 0 {
					b.WriteString(styWarn.Render("backup: "+strings.Join(r.Stow.BackupRels, ", ")) + "\n")
				}
			}
		}
	case tabValidate:
		if m.validation != nil {
			b.WriteString("target: " + m.validation.Target + "\n")
			for _, tool := range m.validation.Tools {
				b.WriteString("  " + tool.Tool + "  " + statusLabel(tool.Status) + "  " + tool.Detail + "\n")
			}
			for _, config := range m.validation.Configs {
				b.WriteString("  " + config.Config.Name + "  " + statusLabel(config.Status) + "  " + config.Detail + "\n")
			}
		} else {
			b.WriteString("Enter to run post-install validation.\n")
		}
	case tabTheme:
		b.WriteString("Themes rewrite tracked config files in place.\n\n")
		b.WriteString("Enter applies the selected palette.\n")
	case tabSummary:
		b.WriteString(m.summaryText())
	}
	return b.String()
}

func (m *model) summaryText() string {
	var b strings.Builder
	if m.validation == nil && m.themeName == "" {
		b.WriteString("No operation has completed yet.\n\nSelect Tools or Configs, then press enter.\n")
		return b.String()
	}
	if m.validation != nil {
		if m.validation.Valid {
			b.WriteString(styOK.Render("POST-VALIDATION: PASS") + "\n")
		} else {
			b.WriteString(styFail.Render("POST-VALIDATION: FAIL") + "\n")
		}
		b.WriteString("target: " + m.validation.Target + "\n")
	}
	if m.themeName != "" {
		b.WriteString("theme: " + m.themeName + "\n")
		if m.themeErr != nil {
			b.WriteString(styFail.Render(m.themeErr.Error()) + "\n")
		}
	}
	return b.String()
}

func statusLabel(status PreflightStatus) string {
	switch status {
	case PreflightInstalled:
		return styOK.Render("installed")
	case PreflightAvailable:
		return styAccent.Render("available")
	case PreflightMissing:
		return styWarn.Render("missing")
	case PreflightConflict:
		return styWarn.Render("conflict / backup")
	case PreflightUnavailable:
		return styFail.Render("unavailable")
	case PreflightUnsupported:
		return styFail.Render("unsupported")
	case PreflightDangling:
		return styWarn.Render("dangling")
	default:
		return string(status)
	}
}

func statusIcon(status stepStatus) string {
	switch status {
	case stOK:
		return styOK.Render("✓")
	case stFail:
		return styFail.Render("✗")
	case stRunning:
		return styRun.Render("…")
	case stSkip:
		return styDim.Render("−")
	default:
		return styDim.Render("•")
	}
}

func (m *model) View() string {
	if m.screen == scrConfigTarget {
		return m.header() + "\n" + styBold.Render("Choose stow target") + "\n\n  " + m.targetInput.View() + "\n\n" + styDim.Render("enter re-plan and continue · esc back · q quit") + "\n"
	}
	if m.screen == scrDone {
		return m.header() + "\n" + styTitle.Render("Summary") + "\n\n" + m.summaryText() + "\n\n" + styDim.Render("enter/esc dashboard · q quit") + "\n"
	}
	var b strings.Builder
	b.WriteString(m.header() + "\n")
	b.WriteString(m.tabs() + "\n")
	if m.working {
		b.WriteString(styRun.Render(m.spin.View()+" "+m.workLabel) + "\n\n")
	}
	if m.width < 90 {
		b.WriteString(m.singleColumn())
	} else {
		b.WriteString(m.dashboardColumns())
	}
	b.WriteString("\n" + m.help())
	return b.String()
}

func (m *model) header() string {
	distro := m.system.Distro.String()
	manager := string(m.system.Kind())
	if manager == "" {
		manager = "unsupported"
	}
	return styTitle.Render("dotfiles-install") + "  " + styDim.Render(distro+" · "+manager+" · "+m.system.Arch) + "  " + styDim.Render(m.target)
}

func (m *model) tabs() string {
	var parts []string
	for i, name := range tabNames {
		if !m.tabAllowed(dashboardTab(i)) {
			continue
		}
		if dashboardTab(i) == m.tab {
			parts = append(parts, styAccent.Bold(true).Render("["+name+"]"))
		} else {
			parts = append(parts, styDim.Render(" "+name+" "))
		}
	}
	return strings.Join(parts, "  ") + "\n" + styRule.Render(strings.Repeat("─", minInt(m.width, 120)))
}

func (m *model) dashboardColumns() string {
	left := m.tableText()
	right := m.detailPane()
	lw := m.width/2 - 3
	if lw < 30 {
		lw = 30
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(lw).Render(left), "  "+styRule.Render("│")+"  ", lipgloss.NewStyle().Width(lw).Render(right))
}

func (m *model) detailPane() string {
	return m.detail.View() + "\n" + styRule.Render("Logs / warnings") + "\n" + m.log.View()
}

func (m *model) singleColumn() string {
	return m.tableText() + "\n" + styRule.Render("Details") + "\n" + m.detailPane()
}

func (m *model) tableText() string {
	var b strings.Builder
	switch m.tab {
	case tabTools:
		b.WriteString(styBold.Render("Tools") + "  " + m.counts() + "\n\n")
		for i, tool := range m.manifest.Tools {
			cursor := "  "
			if i == m.cursor {
				cursor = styAccent.Render("› ")
			}
			check := "[ ]"
			if m.toolSelected[i] {
				check = "[x]"
			}
			status := ""
			if i < len(m.preflight.ToolResults) {
				status = "  " + statusLabel(m.preflight.ToolResults[i].Status)
			}
			b.WriteString(cursor + check + " " + fit(tool.Name, 16) + status + "\n")
		}
	case tabConfigs:
		b.WriteString(styBold.Render("Configs") + "  " + m.counts() + "\n\n")
		for i, pkg := range m.pkgs {
			cursor := "  "
			if i == m.cursor {
				cursor = styAccent.Render("› ")
			}
			check := "[ ]"
			if m.pkgSelected[i] {
				check = "[x]"
			}
			status := ""
			if i < len(m.preflight.ConfigPlans) {
				status = "  " + statusLabel(m.preflight.ConfigPlans[i].Status)
			}
			b.WriteString(cursor + check + " " + fit(pkg.Name, 20) + status + "\n")
		}
	case tabValidate:
		b.WriteString(styBold.Render("Validate") + "\n\n")
		if m.validation == nil {
			b.WriteString("Press enter to validate selected tools/configs.\n")
		} else if m.validation.Valid {
			b.WriteString(styOK.Render("PASS") + "  all selected items are installed\n")
		} else {
			b.WriteString(styFail.Render("FAIL") + "  inspect details and warnings →\n")
		}
	case tabTheme:
		b.WriteString(styBold.Render("Theme") + "\n\n")
		for i, name := range themeDisplayNames {
			cursor := "  "
			if i == m.cursor {
				cursor = styAccent.Render("› ")
			}
			b.WriteString(cursor + name + "\n")
		}
	case tabSummary:
		b.WriteString(styBold.Render("Summary") + "\n\n" + m.summaryText() + "\n")
	}
	return b.String()
}

func (m *model) counts() string {
	selected := 0
	available := 0
	for i, yes := range m.toolSelected {
		if yes {
			selected++
		}
		if i < len(m.preflight.ToolResults) && m.preflight.ToolResults[i].Status == PreflightAvailable {
			available++
		}
	}
	for i, yes := range m.pkgSelected {
		if yes {
			selected++
		}
		if i < len(m.preflight.ConfigPlans) && m.preflight.ConfigPlans[i].Status == PreflightAvailable {
			available++
		}
	}
	return fmt.Sprintf("selected %d · ready %d", selected, available)
}

func (m *model) help() string {
	if m.working {
		return styDim.Render("working… navigation locked")
	}
	return styDim.Render("j/k ↑/↓ move · h/l ←/→ tabs · space toggle · a all · enter run · r refresh · esc back · q quit")
}

func fit(s string, width int) string {
	if width <= 1 {
		return ""
	}
	if len([]rune(s)) > width {
		return string([]rune(s)[:width-1]) + "…"
	}
	return s + strings.Repeat(" ", width-len([]rune(s)))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func runTUI(m *Manifest, pkgs []StowPkg, repoRoot string) int {
	// Config-only runs must not prompt for root before the dashboard is visible.
	// Tool execution uses sudo -n and reports a clear retryable error when the
	// caller has not authenticated yet; an explicit headless install still
	// prepares credentials in main.go before mutating the system.
	if sudoReady() {
		go keepSudoWarm()
	}

	ui := newModel(m, pkgs, repoRoot)
	p := tea.NewProgram(ui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TUI error:", err)
		return 1
	}
	if ui.failures {
		return 1
	}
	return 0
}
