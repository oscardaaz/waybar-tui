package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"waybar-tui/internal/installer"
	"waybar-tui/internal/theme"
)

const leftPanelW = 26 // total width including border

type appState int

const (
	stateMain    appState = iota
	stateInstall
	stateConfirm
)

type refreshMsg struct{}

// Model is the root bubbletea model.
type Model struct {
	state  appState
	width  int
	height int

	themes      []string
	cursor      int
	activeTheme string
	previewTab  int
	preview     [2]string
	vp          viewport.Model
	vpReady     bool

	status    string
	statusErr bool

	install      installModel
	confirmTheme string
}

func New() Model {
	return Model{install: newInstallModel()}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg { return refreshMsg{} }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpW := m.width - leftPanelW - 4
		vpH := m.height - 6
		if vpW < 1 {
			vpW = 1
		}
		if vpH < 1 {
			vpH = 1
		}
		if !m.vpReady {
			m.vp = viewport.New(vpW, vpH)
			m.vpReady = true
		} else {
			m.vp.Width = vpW
			m.vp.Height = vpH
		}
		return m, nil

	case refreshMsg:
		themes, _ := theme.ListThemes()
		m.themes = themes
		m.activeTheme = theme.GetActive()
		if m.cursor >= len(m.themes) && len(m.themes) > 0 {
			m.cursor = len(m.themes) - 1
		}
		m.loadPreview()
		return m, nil

	case cloneDoneMsg:
		if m.state == stateInstall {
			cmd := m.install.handleCloneDone(msg)
			return m, cmd
		}
		return m, nil

	case installDoneMsg:
		if msg.err != nil {
			m.install.errMsg = msg.err.Error()
			m.install.step = stepName
			m.install.name.Focus()
			return m, nil
		}
		m.state = stateMain
		m.status = "✓ Theme '" + msg.name + "' installed"
		m.statusErr = false
		return m, func() tea.Msg { return refreshMsg{} }
	}

	switch m.state {
	case stateMain:
		return m.updateMain(msg)
	case stateInstall:
		return m.updateInstall(msg)
	case stateConfirm:
		return m.updateConfirm(msg)
	}
	return m, nil
}

func (m Model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)

	if isKey {
		m.status = ""
		switch key.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.loadPreview()
			}
		case "down", "j":
			if m.cursor < len(m.themes)-1 {
				m.cursor++
				m.loadPreview()
			}
		case "enter":
			if len(m.themes) == 0 {
				break
			}
			name := m.themes[m.cursor]
			if err := theme.Apply(name); err != nil {
				m.status = "✗ " + err.Error()
				m.statusErr = true
			} else {
				m.activeTheme = name
				m.status = "✓ Applied: " + name
				m.statusErr = false
				return m, func() tea.Msg { return refreshMsg{} }
			}
		case "i":
			m.install = newInstallModel()
			m.state = stateInstall
		case "d":
			if len(m.themes) > 0 {
				m.confirmTheme = m.themes[m.cursor]
				m.state = stateConfirm
			}
		case "r":
			return m, func() tea.Msg { return refreshMsg{} }
		case "tab":
			m.previewTab = 1 - m.previewTab
			m.vp.SetContent(m.preview[m.previewTab])
			m.vp.GotoTop()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) updateInstall(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		installer.Cleanup(m.install.tmpPath)
		m.state = stateMain
		return m, nil
	}
	cmd := m.install.update(msg)
	return m, cmd
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "y":
		if err := theme.Delete(m.confirmTheme); err != nil {
			m.status = "✗ " + err.Error()
			m.statusErr = true
		} else {
			m.status = "✓ Theme '" + m.confirmTheme + "' deleted"
			m.statusErr = false
		}
		m.state = stateMain
		return m, func() tea.Msg { return refreshMsg{} }
	case "n", "esc":
		m.state = stateMain
	}
	return m, nil
}

func (m *Model) loadPreview() {
	if len(m.themes) == 0 || !m.vpReady {
		return
	}
	files := theme.GetFiles(m.themes[m.cursor])
	m.preview[0] = files["config.jsonc"]
	m.preview[1] = files["style.css"]
	m.vp.SetContent(m.preview[m.previewTab])
	m.vp.GotoTop()
}

// ── Views ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if !m.vpReady {
		return ""
	}
	switch m.state {
	case stateMain:
		return m.viewMain()
	case stateInstall:
		return m.viewInstall()
	case stateConfirm:
		return m.viewConfirm()
	}
	return ""
}

func (m Model) viewMain() string {
	panelH := m.height - 2

	// Left panel
	left := styleBorder.
		Width(leftPanelW - 2).
		Height(panelH - 2).
		Render(m.renderList(panelH - 2))

	// Right panel
	rightW := m.width - leftPanelW - 2
	right := styleBorder.
		Width(rightW - 2).
		Height(panelH - 2).
		Render(m.renderTabs() + "\n" + m.vp.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + m.renderHelp()
}

func (m Model) renderList(maxLines int) string {
	if len(m.themes) == 0 {
		return styleMuted.Render("No themes\npress i to install")
	}
	start := 0
	if m.cursor >= maxLines {
		start = m.cursor - maxLines + 1
	}
	var sb strings.Builder
	for i := start; i < len(m.themes) && i < start+maxLines; i++ {
		name := m.themes[i]
		maxNameW := leftPanelW - 6
		if len(name) > maxNameW {
			name = name[:maxNameW-1] + "…"
		}
		var line string
		isSelected := i == m.cursor
		isActive := m.themes[i] == m.activeTheme
		switch {
		case isSelected && isActive:
			line = styleSelected.Render("▶ " + name + " ●")
		case isSelected:
			line = styleSelected.Render("▶ " + name)
		case isActive:
			line = styleActive.Render("  " + name + " ●")
		default:
			line = "  " + name
		}
		if i > start {
			sb.WriteString("\n")
		}
		sb.WriteString(line)
	}
	return sb.String()
}

func (m Model) renderTabs() string {
	tabs := [2]string{"config.jsonc", "style.css"}
	var parts []string
	for i, t := range tabs {
		if i == m.previewTab {
			parts = append(parts, styleTabActive.Render(t))
		} else {
			parts = append(parts, styleTabInactive.Render(t))
		}
	}
	return strings.Join(parts, "")
}

func (m Model) renderHelp() string {
	sep := styleMuted.Render("  ·  ")
	keys := strings.Join([]string{
		helpKey("↑↓", "navigate"),
		helpKey("enter", "apply"),
		helpKey("i", "install"),
		helpKey("d", "delete"),
		helpKey("tab", "preview"),
		helpKey("q", "quit"),
	}, sep)

	if m.status != "" {
		st := styleSuccess.Render(m.status)
		if m.statusErr {
			st = styleError.Render(m.status)
		}
		return "  " + st + sep + keys
	}
	return "  " + keys
}

func (m Model) viewInstall() string {
	var body strings.Builder

	body.WriteString(styleMuted.Render("Repository URL") + "\n")
	body.WriteString(m.install.url.View() + "\n\n")

	if m.install.errMsg != "" {
		body.WriteString(styleError.Render("✗ "+m.install.errMsg) + "\n\n")
	} else if m.install.status != "" {
		body.WriteString(styleSuccess.Render("→ "+m.install.status) + "\n\n")
	} else if m.install.step == stepCloning {
		body.WriteString(styleMuted.Render("Cloning...") + "\n\n")
	} else {
		body.WriteString("\n")
	}

	switch m.install.step {
	case stepSelect:
		body.WriteString(styleMuted.Render("Filter") + "\n")
		body.WriteString(m.install.filter.View() + "\n\n")
		body.WriteString(m.install.viewCandidates() + "\n\n")
		body.WriteString(styleMuted.Render("Theme name") + "\n")
		body.WriteString(m.install.name.View() + "\n")
	case stepName:
		body.WriteString(styleMuted.Render("Theme name") + "\n")
		body.WriteString(m.install.name.View() + "\n")
	}

	body.WriteString("\n" + helpKey("enter", "confirm") + styleMuted.Render("  ·  ") + helpKey("esc", "cancel"))

	modal := styleModal.Width(60).Render(
		styleTitle.Render("Install from GitHub") + "\n\n" + body.String(),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) viewConfirm() string {
	content := styleTitle.Render("Delete theme") + "\n\n" +
		"  Delete '" + m.confirmTheme + "'?\n" +
		"  This action cannot be undone.\n\n" +
		"  " + helpKey("y", "confirm") + styleMuted.Render("  ·  ") + helpKey("n / esc", "cancel")

	modal := styleModal.Width(44).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}
