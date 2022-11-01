package view

import (
	"flightcrew.io/cli/internal/view/command"
	tea "github.com/charmbracelet/bubbletea"
)

type endModel struct {
	// All of the commands that were run that should be displayed and maybe printed.
	commands []*command.Model
}

func NewEndModel() endModel {
	return endModel{}
}

func (m *endModel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m endModel) View() string {
	return ""
}
