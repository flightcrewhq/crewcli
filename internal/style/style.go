package style

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	Focused = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	Blurred = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	None    = lipgloss.NewStyle()
	Code    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CA5D8"))
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D88CA5")).Bold(true).Render
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CD8B2")).Bold(true).Render
	Bold    = lipgloss.NewStyle().Bold(true).Render
	Help    = lipgloss.NewStyle().Padding(1, 0, 0, 2).Foreground(lipgloss.AdaptiveColor{
		Light: "#737675",
		Dark:  "#D8DAD9",
	}).Render

	Glamour *glamour.TermRenderer
)

func init() {
	var err error
	Glamour, err = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(80),
	)

	if err != nil {
		panic(err)
	}
}
