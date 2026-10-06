package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"gopkg.in/yaml.v3"
	"os"
)

type row struct {
	Category string
	Index    int
}
type picker struct {
	catalog               Catalog
	profile               Profile
	selected              map[string]bool
	rows                  []row
	cursor, height, width int
	review, accepted      bool
	message               string
}

var nordTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#88c0d0")).Bold(true)
var nordSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("#eceff4")).Background(lipgloss.Color("#434c5e"))

func newPicker(c Catalog, p Profile) picker {
	m := picker{catalog: c, profile: p, selected: map[string]bool{}, height: 24, width: 100}
	chosen, _ := selectTools(c, p)
	for _, t := range chosen {
		m.selected[t.Name] = true
	}
	cats := []string{}
	seen := map[string]bool{}
	for _, t := range c.Tools {
		if !seen[t.Category] {
			cats = append(cats, t.Category)
			seen[t.Category] = true
		}
	}
	for _, cat := range cats {
		m.rows = append(m.rows, row{Category: cat, Index: -1})
		for i, t := range c.Tools {
			if t.Category == cat {
				m.rows = append(m.rows, row{Category: cat, Index: i})
			}
		}
	}
	return m
}
func (m picker) Init() tea.Cmd { return nil }
func (m picker) selection() Profile {
	p := m.profile
	p.Tools = nil
	p.Categories = nil
	p.Exclude = nil
	for _, t := range m.catalog.Tools {
		if m.selected[t.Name] {
			p.Tools = append(p.Tools, t.Name)
		}
	}
	return p
}
func (m picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		if m.review {
			if key == "esc" || key == "backspace" {
				m.review = false
			}
			if key == "enter" {
				m.accepted = true
				return m, tea.Quit
			}
			return m, nil
		}
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "pgdown":
			m.cursor = min(len(m.rows)-1, m.cursor+max(1, m.height-10))
		case "pgup":
			m.cursor = max(0, m.cursor-max(1, m.height-10))
		case " ":
			r := m.rows[m.cursor]
			if r.Index >= 0 {
				t := m.catalog.Tools[r.Index]
				m.selected[t.Name] = !m.selected[t.Name]
			} else {
				all := true
				for _, t := range m.catalog.Tools {
					if t.Category == r.Category && !m.selected[t.Name] {
						all = false
					}
				}
				for _, t := range m.catalog.Tools {
					if t.Category == r.Category {
						m.selected[t.Name] = !all
					}
				}
			}
		case "a":
			all := true
			for _, t := range m.catalog.Tools {
				if !m.selected[t.Name] {
					all = false
				}
			}
			for _, t := range m.catalog.Tools {
				m.selected[t.Name] = !all
			}
		case "c":
			m.profile.Configs = !m.profile.Configs
		case "i":
			m.profile.Install = !m.profile.Install
		case "m":
			if m.profile.ConfigMode == "copy" {
				m.profile.ConfigMode = "link"
			} else {
				m.profile.ConfigMode = "copy"
			}
		case "b":
			if m.profile.Conflict == "backup" {
				m.profile.Conflict = "error"
			} else {
				m.profile.Conflict = "backup"
			}
		case "s":
			data, err := yaml.Marshal(m.selection())
			if err == nil {
				var f *os.File
				f, err = os.OpenFile("dotfiles-selection.yaml", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
				if err == nil {
					_, err = f.Write(data)
					closeErr := f.Close()
					if err == nil {
						err = closeErr
					}
				}
			}
			if err != nil {
				m.message = err.Error()
			} else {
				m.message = "Saved dotfiles-selection.yaml"
			}
		case "enter":
			if len(m.selection().Tools) == 0 && len(m.profile.Packages) == 0 {
				m.message = "Select tools or packages first"
			} else if !m.profile.Install && !m.profile.Configs {
				m.message = "Enable installation or configs"
			} else {
				m.review = true
			}
		}
	}
	return m, nil
}
func (m picker) View() string {
	var b strings.Builder
	b.WriteString(nordTitle.Render("DOTFILES / NORD") + "\n")
	fmt.Fprintf(&b, "[i] install: %t   [c] configs: %t   [m] %s   [b] conflicts: %s\n", m.profile.Install, m.profile.Configs, m.profile.ConfigMode, m.profile.Conflict)
	if m.review {
		p := m.selection()
		b.WriteString("\nApply selection:\n" + strings.Join(p.Tools, ", ") + "\n")
		if len(p.Packages) > 0 {
			b.WriteString("Extra native packages: " + strings.Join(p.Packages, ", ") + "\n")
		}
		b.WriteString("\nEnter: apply   Esc: back   q: cancel\n")
		b.WriteString("Existing configs are backed up unless conflicts=error.\nPackage installation may require sudo.\n")
		return lipgloss.NewStyle().MaxWidth(max(20, m.width)).Render(b.String())
	}
	visible := max(1, m.height-8)
	start := max(0, m.cursor-visible+1)
	end := min(len(m.rows), start+visible)
	for _, r := range m.rows[start:end] {
		text := ""
		if r.Index < 0 {
			n, total := 0, 0
			for _, t := range m.catalog.Tools {
				if t.Category == r.Category {
					total++
					if m.selected[t.Name] {
						n++
					}
				}
			}
			mark := " "
			if n == total {
				mark = "x"
			} else if n > 0 {
				mark = "-"
			}
			text = fmt.Sprintf("[%s] %s (%d/%d)", mark, r.Category, n, total)
		} else {
			t := m.catalog.Tools[r.Index]
			mark := " "
			if m.selected[t.Name] {
				mark = "x"
			}
			suffix := ""
			if t.Config != "" {
				suffix = " + config"
			}
			if t.External {
				suffix += " [external]"
			}
			text = fmt.Sprintf("    [%s] %-14s %s%s", mark, t.Name, t.Description, suffix)
		}
		if m.rows[m.cursor] == r {
			text = nordSelected.Render(text)
		}
		b.WriteString(text + "\n")
	}
	b.WriteString("\n↑/↓ j/k: move · Space: tool/category · a: all/none\nEnter: review · s: save YAML · q: cancel\n")
	if len(m.profile.Packages) > 0 {
		b.WriteString("Extra packages: " + strings.Join(m.profile.Packages, ", ") + "\n")
	}
	b.WriteString(m.message)
	return lipgloss.NewStyle().MaxWidth(max(20, m.width)).Render(b.String())
}
func choose(c Catalog, p Profile) (Profile, error) {
	if !term.IsTerminal(os.Stdin.Fd()) {
		return p, fmt.Errorf("TUI requires a terminal; use --profile ... --yes or --dry-run")
	}
	model, err := tea.NewProgram(newPicker(c, p), tea.WithAltScreen()).Run()
	if err != nil {
		return p, err
	}
	m := model.(picker)
	if !m.accepted {
		return Profile{}, nil
	}
	return m.selection(), nil
}
