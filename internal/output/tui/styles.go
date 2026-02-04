package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title       lipgloss.Style
	TestPassed  lipgloss.Style
	TestFailed  lipgloss.Style
	TestRunning lipgloss.Style
	TestSkipped lipgloss.Style
	ErrorMsg    lipgloss.Style
	ErrorDetail lipgloss.Style
	HelpBar     lipgloss.Style
	Cursor      lipgloss.Style
	Panel       lipgloss.Style
	ActivePanel lipgloss.Style
	Dim         lipgloss.Style
	Bold        lipgloss.Style

	IconRunning string
	IconExpand  string
	IconCollaps string
}

func DefaultStyles() Styles {
	green := lipgloss.Color("2")
	red := lipgloss.Color("1")
	yellow := lipgloss.Color("3")
	cyan := lipgloss.Color("6")
	dim := lipgloss.Color("8")
	white := lipgloss.Color("15")
	blue := lipgloss.Color("4")

	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(blue).
			Padding(0, 1),

		TestPassed: lipgloss.NewStyle().
			Foreground(green),

		TestFailed: lipgloss.NewStyle().
			Foreground(red),

		TestRunning: lipgloss.NewStyle().
			Foreground(cyan),

		TestSkipped: lipgloss.NewStyle().
			Foreground(yellow),

		ErrorMsg: lipgloss.NewStyle().
			Foreground(yellow),

		ErrorDetail: lipgloss.NewStyle().
			Foreground(dim),

		HelpBar: lipgloss.NewStyle().
			Foreground(dim),

		Cursor: lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(white),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dim).
			Padding(0, 1),

		ActivePanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cyan).
			Padding(0, 1),

		Dim: lipgloss.NewStyle().
			Foreground(dim),

		Bold: lipgloss.NewStyle().
			Bold(true),

		IconRunning: "●",
		IconExpand:  "▼",
		IconCollaps: "►",
	}
}
