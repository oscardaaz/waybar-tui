package ui

import "github.com/charmbracelet/lipgloss"

// Use ANSI colors 0-15 so the UI automatically follows the terminal theme
// (catppuccin, tokyo night, gruvbox, etc.)
var (
	colorAccent   = lipgloss.Color("4")  // blue
	colorGreen    = lipgloss.Color("2")  // green
	colorMuted    = lipgloss.Color("8")  // bright black
	colorError    = lipgloss.Color("1")  // red
	colorSelected = lipgloss.Color("12") // bright blue (background for selection)
	colorBorder   = lipgloss.Color("8")  // bright black

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	styleSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colorSelected).
			Bold(true)

	styleActive = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleAccent = lipgloss.NewStyle().
			Foreground(colorAccent)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleTabActive = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Underline(true).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1)

	styleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)
)

func helpKey(key, desc string) string {
	return styleAccent.Render(key) + " " + styleMuted.Render(desc)
}
