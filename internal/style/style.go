package style

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const defaultWidth = 80

var (
	Focused = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	Blurred = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	None    = lipgloss.NewStyle()
	Code    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CA5D8"))
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D88CA5")).Bold(true).Render
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CD8B2")).Bold(true).Render
	Bold    = lipgloss.NewStyle().Bold(true).Render
	Help    = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).PaddingTop(1).Width(defaultWidth).Foreground(lipgloss.AdaptiveColor{
		Light: "#737675",
		Dark:  "240",
	}).Render
	Action    = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF7EFF")).Bold(true).Render
	Required  = lipgloss.NewStyle().Foreground(lipgloss.Color("199")).Render
	Highlight = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
		Light: "205",
		Dark:  "205",
	}).Foreground(lipgloss.AdaptiveColor{
		Light: "#FFFFFF",
		Dark:  "#FFFFFF",
	}).Render

	Glamour *glamour.TermRenderer
)

func init() {
	var err error
	Glamour, err = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(defaultWidth),
	)

	if err != nil {
		panic(err)
	}
}
