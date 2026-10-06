package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed tools.yaml
var catalogYAML []byte

type Release struct {
	VersionURL  string            `yaml:"version_url"`
	URL         string            `yaml:"url"`
	ChecksumURL string            `yaml:"checksum_url"`
	Raw         bool              `yaml:"raw"`
	Repo        string            `yaml:"repo"`
	Assets      map[string]string `yaml:"assets"`
}
type Tool struct {
	Name        string              `yaml:"name"`
	Category    string              `yaml:"category"`
	Description string              `yaml:"description"`
	Bin         string              `yaml:"bin"`
	Config      string              `yaml:"config"`
	Packages    map[string][]string `yaml:"packages"`
	Release     *Release            `yaml:"release"`
	External    bool                `yaml:"external"`
}
type Catalog struct {
	Tools []Tool `yaml:"tools"`
}
type Profile struct {
	Version    int      `yaml:"version"`
	Categories []string `yaml:"categories,omitempty"`
	Tools      []string `yaml:"tools,omitempty"`
	Exclude    []string `yaml:"exclude,omitempty"`
	Packages   []string `yaml:"packages,omitempty"`
	Install    bool     `yaml:"install"`
	Configs    bool     `yaml:"configs"`
	ConfigMode string   `yaml:"config_mode"`
	Conflict   string   `yaml:"conflict"`
}

func defaultProfile() Profile {
	return Profile{Version: 1, Install: true, Configs: true, ConfigMode: "copy", Conflict: "backup"}
}
func decodeStrict(data []byte, into any) error {
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(into); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected exactly one YAML/JSON document")
	}
	return nil
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9+._:-]*$`)

func loadCatalog() (Catalog, error) {
	var c Catalog
	if err := decodeStrict(catalogYAML, &c); err != nil {
		return c, err
	}
	seen := map[string]bool{}
	for _, t := range c.Tools {
		if !validName.MatchString(t.Name) || seen[t.Name] || t.Category == "" {
			return c, fmt.Errorf("invalid/duplicate tool %q", t.Name)
		}
		seen[t.Name] = true
		if t.Config != "" && (!validName.MatchString(t.Config) || strings.Contains(t.Config, "..")) {
			return c, fmt.Errorf("invalid config: %s", t.Config)
		}
		if t.Release != nil {
			for _, pattern := range t.Release.Assets {
				if _, err := regexp.Compile(pattern); err != nil {
					return c, err
				}
			}
		}
	}
	return c, nil
}
func readProfile(path string) (Profile, error) {
	p := defaultProfile()
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	err = decodeStrict(data, &p)
	return p, err
}
func selectTools(c Catalog, p Profile) ([]Tool, error) {
	if p.Version != 1 {
		return nil, fmt.Errorf("unsupported profile version %d", p.Version)
	}
	if p.ConfigMode != "copy" && p.ConfigMode != "link" {
		return nil, fmt.Errorf("config_mode must be copy or link")
	}
	if p.Conflict != "backup" && p.Conflict != "error" {
		return nil, fmt.Errorf("conflict must be backup or error")
	}
	names, categories := map[string]bool{}, map[string]bool{}
	for _, t := range c.Tools {
		names[t.Name] = true
		categories[t.Category] = true
	}
	chosen := map[string]bool{}
	for _, cat := range p.Categories {
		if !categories[cat] {
			return nil, fmt.Errorf("unknown category %q", cat)
		}
		for _, t := range c.Tools {
			if t.Category == cat {
				chosen[t.Name] = true
			}
		}
	}
	for _, name := range p.Tools {
		if !names[name] {
			return nil, fmt.Errorf("unknown tool %q", name)
		}
		chosen[name] = true
	}
	for _, name := range p.Exclude {
		if !names[name] {
			return nil, fmt.Errorf("unknown excluded tool %q", name)
		}
		delete(chosen, name)
	}
	for _, pkg := range p.Packages {
		if !validName.MatchString(pkg) {
			return nil, fmt.Errorf("invalid package name %q (use native package names without options/versions)", pkg)
		}
	}
	var out []Tool
	for _, t := range c.Tools {
		if chosen[t.Name] {
			out = append(out, t)
		}
	}
	return out, nil
}
