package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
	"testing"
)

func TestProfileSelectionAndValidation(t *testing.T) {
	c, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	p := defaultProfile()
	p.Categories = []string{"git"}
	p.Tools = []string{"fish", "git"}
	p.Exclude = []string{"gh"}
	selected, err := selectTools(c, p)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, tool := range selected {
		if seen[tool.Name] {
			t.Fatal("duplicate")
		}
		seen[tool.Name] = true
	}
	if !seen["git"] || !seen["fish"] || !seen["lazygit"] || seen["gh"] {
		t.Fatal(seen)
	}
	for _, test := range []struct{ field, value string }{{"tools", "oops"}, {"categories", "oops"}, {"exclude", "oops"}, {"packages", "--evil"}, {"packages", "git;touch /tmp/x"}, {"config_mode", "oops"}, {"conflict", "overwrite"}} {
		t.Run(test.field+test.value, func(t *testing.T) {
			p := defaultProfile()
			switch test.field {
			case "tools":
				p.Tools = []string{test.value}
			case "categories":
				p.Categories = []string{test.value}
			case "exclude":
				p.Exclude = []string{test.value}
			case "packages":
				p.Packages = []string{test.value}
			case "config_mode":
				p.ConfigMode = test.value
			case "conflict":
				p.Conflict = test.value
			}
			if _, err := selectTools(c, p); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
func TestStrictProfileAndJSON(t *testing.T) {
	for _, data := range []string{"version: 1\nconfigz: false\n", "version: 1\n---\nversion: 1\n", "version: 1\nversion: 2\n", "tools: fish\n"} {
		p := defaultProfile()
		if err := decodeStrict([]byte(data), &p); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	p := defaultProfile()
	if err := decodeStrict([]byte(`{"version":1,"tools":["fish"],"configs":false}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.Configs || !p.Install || p.ConfigMode != "copy" {
		t.Fatal(p)
	}
}
func TestExamples(t *testing.T) {
	c, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"vps.yaml", "workstation.yaml", "config-only.json"} {
		p, err := readProfile(filepath.Join("examples", name))
		if err != nil {
			t.Fatal(err)
		}
		tools, err := selectTools(c, p)
		if err != nil {
			t.Fatal(err)
		}
		if len(tools) == 0 {
			t.Fatal("empty example")
		}
		for _, tool := range tools {
			if tool.Config != "" {
				if _, err := os.Stat(filepath.Join("..", "config", tool.Config)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
func TestPickerCategoryAndModes(t *testing.T) {
	c, _ := loadCatalog()
	p := defaultProfile()
	m := newPicker(c, p)
	if len(m.selection().Tools) != 0 {
		t.Fatal("default must be empty")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(picker)
	for _, tool := range c.Tools {
		if tool.Category == m.rows[0].Category && !m.selected[tool.Name] {
			t.Fatal(tool.Name)
		}
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(picker)
	if m.profile.Configs {
		t.Fatal("config toggle")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(picker)
	if !m.review || m.accepted {
		t.Fatal("must review before applying")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(picker)
	if m.review {
		t.Fatal("back should keep selection")
	}
}
