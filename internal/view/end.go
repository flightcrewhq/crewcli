package view

import (
	"fmt"
	"strings"
	"time"

	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"flightcrew.io/cli/internal/view/wrapinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EndModel struct {
	inputs Inputs

	// All of the commands that were run that should be displayed and maybe printed.
	commands []*command.Model

	writeButton *button.Button
	writeInput  wrapinput.Model
	wrote       bool
}

func NewEndModel(inputs Inputs) *EndModel {
	writeButton, _ := button.New("Write", 10)
	m := &EndModel{
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

func (m *EndModel) Init() tea.Cmd {
	return nil
}

func (m *EndModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m EndModel) View() string {
	var b strings.Builder
	b.WriteString(m.inputs.EndDescription())
	b.WriteRune('\n')
	b.WriteString("Print output of commands to file? ")
	b.WriteString(m.writeInput.View(wrapinput.ViewParams{ShowValue: false}))
	return b.String()
}
