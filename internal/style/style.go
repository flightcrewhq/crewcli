package style

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	Focused = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	Blurred = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	None    = lipgloss.NewStyle()
	Code    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CA5D8"))
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D88CA5")).Render
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CD8B2")).Render

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
