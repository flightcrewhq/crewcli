package view

import (
	"fmt"
	"time"

	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"flightcrew.io/cli/internal/view/wrapinput"
	tea "github.com/charmbracelet/bubbletea"
)

type endModel struct {
	inputs Inputs

	// All of the commands that were run that should be displayed and maybe printed.
	commands []*command.Model

	writeButton *button.Button
	writeInput  wrapinput.Model
	wrote       bool
}

func NewEndModel(inputs Inputs) endModel {
	writeButton, _ := button.New("Write", 10)
	m := endModel{
		inputs:      inputs,
		commands:    inputs.Commands(),
		writeButton: writeButton,
		writeInput:  wrapinput.NewFreeForm(),
		wrote:       false,
	}

	input := m.writeInput.Freeform
	defaultFile := fmt.Sprintf("/tmp/%s_%d/output/log", inputs.Name(), time.Now().Unix())
	input.Placeholder = defaultFile
	m.writeInput.Default = defaultFile

	return m
}

func (m *endModel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m endModel) View() string {
	return ""
}
