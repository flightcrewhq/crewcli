package view

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"flightcrew.io/cli/internal/view/wrapinput"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EndModel struct {
	controller controller.End

	// All of the commands that were run that should be displayed and maybe printed.
	commands []*command.Model

	yesButton  *button.Button
	noButton   *button.Button
	writeInput wrapinput.Model
	wrote      bool
	userInput  bool

	confirming bool
}

func NewEndModel(ctl controller.End) *EndModel {
	yesButton, _ := button.New("Print", 10)
	noButton, _ := button.New("Edit", 10)
	wInput := wrapinput.NewFreeForm()
	wInput.Title = "  Print output of commands to file"
	input := wInput.Freeform
	defaultFile := fmt.Sprintf("/tmp/%s_%d/output/log",
		strings.Replace(strings.ToLower(ctl.Name()), " ", "_", -1),
		time.Now().Unix())
	input.Placeholder = defaultFile
	wInput.Default = defaultFile
	wInput.Focus()

	m := &EndModel{
		controller: ctl,
		commands:   ctl.Commands(),
		yesButton:  yesButton,
		noButton:   noButton,
		writeInput: wInput,
		wrote:      false,
	}

	return m
}

func (m *EndModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *EndModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if !m.confirming {
				m.confirming = true
				m.writeInput.Blur()
				return m, nil
			}

			if m.userInput {
				if err := m.printCommands(m.writeInput.Value()); err != nil {
					m.writeInput.SetError(err)
					return m, nil
				}

				fmt.Printf("\n\nOutput is located at %s\n", m.writeInput.Value())
				return m, tea.Quit
			}

			m.userInput = false
			m.confirming = false
			return m, m.writeInput.Focus()

		case "tab":
			if m.confirming {
				m.userInput = !m.userInput
				return m, nil
			}

		case "left":
			if m.confirming {
				m.userInput = true
				return m, nil
			}

		case "right":
			if m.confirming {
				m.userInput = false
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.writeInput, cmd = m.writeInput.Update(msg)
	return m, cmd
}

func (m EndModel) View() string {
	var b strings.Builder
	b.WriteString(m.controller.EndDescription())
	b.WriteRune('\n')
	b.WriteString(m.writeInput.View(wrapinput.ViewParams{ShowValue: m.confirming}))
	b.WriteRune('\n')
	if m.confirming {
		b.WriteString("  ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')
		b.WriteString(style.Help("ctrl+c/esc: quit • enter: confirm •  ←/→/tab: nav"))
	} else {
		b.WriteRune('\n')
		b.WriteRune('\n')
		b.WriteString(style.Help("ctrl+c/esc: quit • enter: confirm"))
	}

	return b.String()
}

func (m EndModel) printCommands(fn string) error {
	dir, _ := filepath.Split(fn)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	_, _ = f.WriteString(fmt.Sprintf("Output of %s\n\n", m.controller.Name()))

	for _, cmd := range m.commands {
		_, _ = f.WriteString("--------------------\n")
		_, _ = f.WriteString(cmd.String())
		_, _ = f.WriteString("\n")
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
