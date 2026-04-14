package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"waybar-tui/internal/installer"
)

type installStep int

const (
	stepURL installStep = iota
	stepCloning
	stepSelect
	stepName
)

type cloneDoneMsg struct {
	tmpPath    string
	candidates []string
	err        error
}

type installDoneMsg struct {
	name string
	err  error
}

type installModel struct {
	step     installStep
	url      textinput.Model
	filter   textinput.Model
	name     textinput.Model
	tmpPath  string
	all      []string
	filtered []string
	cursor   int
	status   string
	errMsg   string
}

func newInstallModel() installModel {
	url := textinput.New()
	url.Placeholder = "https://github.com/user/repo"
	url.Width = 52
	url.Focus()

	filter := textinput.New()
	filter.Placeholder = "Filter..."
	filter.Width = 52

	name := textinput.New()
	name.Placeholder = "theme-name"
	name.Width = 52

	return installModel{step: stepURL, url: url, filter: filter, name: name}
}

func doClone(url string) tea.Cmd {
	return func() tea.Msg {
		tmp, err := installer.Clone(url)
		if err != nil {
			return cloneDoneMsg{err: err}
		}
		candidates, err := installer.ScanCandidates(tmp)
		return cloneDoneMsg{tmpPath: tmp, candidates: candidates, err: err}
	}
}

func (m *installModel) handleCloneDone(msg cloneDoneMsg) tea.Cmd {
	if msg.err != nil {
		m.errMsg = msg.err.Error()
		m.step = stepURL
		m.url.Focus()
		return nil
	}
	m.tmpPath = msg.tmpPath
	m.all = msg.candidates
	m.applyFilter()

	if len(m.all) == 0 {
		m.errMsg = "No valid themes found in the repository"
		m.step = stepURL
		m.url.Focus()
		return nil
	}
	if len(m.all) == 1 {
		m.name.SetValue(filepath.Base(m.all[0]))
		m.step = stepName
		m.url.Blur()
		m.name.Focus()
		m.status = "1 theme found"
	} else {
		m.step = stepSelect
		m.url.Blur()
		m.filter.Focus()
		m.status = fmt.Sprintf("%d themes found", len(m.all))
	}
	return nil
}

func (m *installModel) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		m.filtered = make([]string, len(m.all))
		copy(m.filtered, m.all)
		return
	}
	m.filtered = m.filtered[:0]
	for _, c := range m.all {
		if strings.Contains(strings.ToLower(filepath.Base(c)), q) {
			m.filtered = append(m.filtered, c)
		}
	}
	if m.cursor >= len(m.filtered) {
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		} else {
			m.cursor = 0
		}
	}
}

func (m *installModel) update(msg tea.Msg) tea.Cmd {
	key, isKey := msg.(tea.KeyMsg)

	switch m.step {
	case stepURL:
		if isKey && key.String() == "enter" {
			if m.url.Value() == "" {
				return nil
			}
			m.step = stepCloning
			m.status = "Cloning..."
			m.errMsg = ""
			return doClone(m.url.Value())
		}
		var cmd tea.Cmd
		m.url, cmd = m.url.Update(msg)
		return cmd

	case stepCloning:
		return nil

	case stepSelect:
		if isKey {
			switch key.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return nil
			case "down", "j":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
				return nil
			case "enter":
				if len(m.filtered) > 0 {
					m.name.SetValue(filepath.Base(m.filtered[m.cursor]))
					m.step = stepName
					m.filter.Blur()
					m.name.Focus()
				}
				return nil
			case "tab":
				return nil
			}
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.applyFilter()
		m.cursor = 0
		return cmd

	case stepName:
		if isKey && key.String() == "enter" {
			if m.name.Value() == "" {
				return nil
			}
			return m.doInstall()
		}
		var cmd tea.Cmd
		m.name, cmd = m.name.Update(msg)
		return cmd
	}

	return nil
}

func (m *installModel) doInstall() tea.Cmd {
	var src string
	if len(m.all) == 1 {
		src = m.all[0]
	} else if m.cursor < len(m.filtered) {
		src = m.filtered[m.cursor]
	} else {
		return nil
	}
	name := m.name.Value()
	tmpPath := m.tmpPath

	return func() tea.Msg {
		err := installer.Install(src, name)
		installer.Cleanup(tmpPath)
		return installDoneMsg{name: name, err: err}
	}
}

func (m installModel) viewCandidates() string {
	if len(m.filtered) == 0 {
		return styleMuted.Render("  No results")
	}
	maxVisible := 8
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	var sb strings.Builder
	for i := start; i < len(m.filtered) && i < start+maxVisible; i++ {
		name := filepath.Base(m.filtered[i])
		if i > start {
			sb.WriteString("\n")
		}
		if i == m.cursor {
			sb.WriteString(styleSelected.Render("  " + name + "  "))
		} else {
			sb.WriteString("  " + name)
		}
	}
	remaining := len(m.filtered) - (start + maxVisible)
	if remaining > 0 {
		sb.WriteString("\n" + styleMuted.Render(fmt.Sprintf("  ... %d more", remaining)))
	}
	return sb.String()
}
