package view

import (
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Inputs is a variable number of inputs that can be changed. This interface allows
// pre-defined defaults that can be modified according to cloud providers' specific
// requirements.
type Inputs interface {
	// Len returns the number of inputs in this grouping.
	Len() int
	// Validate gives the library a chance to convert any values that need to be converted
	// or make sure that all inputs have valid values.
	Validate() bool
	Reset()
	// Args turns the inputs into a map from key to value so that the inputted values can
	// be replaced in the subsequent commands to be run.
	Args() map[string]string

	View() string
	Update(msg tea.Msg) tea.Cmd

	Focus(i int) tea.Cmd
	NextEmpty(i int) int
}

type installModel struct {
	width  int
	height int

	// TODO: Move the GCP stuff into a separate place. Probably group this entire block and turn it into its own model?
	inputs     Inputs
	index      int
	hasErrors  bool
	confirming bool

	submitButton     *button.Button
	confirmYesButton *button.Button
	confirmNoButton  *button.Button

	getCommands LazyCommands

	logStatements []string
}

func NewInstallModel(inputs Inputs, getCommands LazyCommands) installModel {
	debug.Output("new install model!")

	m := installModel{
		inputs:        inputs,
		getCommands:   getCommands,
		logStatements: make([]string, 0),
	}

	m.submitButton, _ = button.New("Submit", 12)
	m.confirmYesButton, _ = button.New("Continue", 12)
	m.confirmNoButton, _ = button.New("Edit", 12)

	m.inputs.Update(nil)
	m.index = m.inputs.NextEmpty(m.index)
	m.inputs.Focus(m.index)

	return m
}

func (m installModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height // height is available too

	case tea.KeyMsg:
		s := msg.String()
		cmds := make([]tea.Cmd, 0)
		switch s {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down", "left", "right":
			switch s {
			case "enter":
				if m.index == m.inputs.Len() {
					if !m.confirming {
						m.confirming = true
						m.hasErrors = !m.inputs.Validate()
						if m.hasErrors {
							m.index = m.inputs.Len() + 1
						}
						return m, nil
					}

					return NewRunModel(m.inputs.Args(), m.getCommands), nil
				}

				if m.index == m.inputs.Len()+1 {
					m.confirming = false
					m.hasErrors = false
					m.index = m.inputs.Len()
					m.inputs.Reset()
					return m, nil
				}

				m.index = m.inputs.NextEmpty(m.index + 1)

			case "up", "shift+tab":
				if !m.confirming {
					m.index--
				}

			case "down":
				if !m.confirming {
					m.index++
				}

			case "tab":
				if m.confirming {
					if m.index == m.inputs.Len()+1 {
						m.index = m.inputs.Len()
					} else if m.index == m.inputs.Len() {
						m.index = m.inputs.Len() + 1
					}
				} else {
					m.index++
				}

			case "left":
				if m.confirming {
					if !m.hasErrors {
						m.index--
					}
				}

			case "right":
				if m.confirming {
					m.index++
				}
			}

			if m.index > m.inputs.Len() {
				if m.confirming && m.index > m.inputs.Len()+1 {
					m.index = m.inputs.Len() + 1
				} else if !m.confirming {
					m.index = 0
				}
			} else if m.confirming && m.index < m.inputs.Len() {
				m.index = m.inputs.Len()
			} else if !m.confirming && m.index < 0 {
				m.index = m.inputs.Len()
			}

			if m.confirming {
				break
			}

			cmds = append(cmds, m.inputs.Focus(m.index))
			if m.index < m.inputs.Len() {
				cmds = append(cmds, m.inputs.Update(msg))
			}
			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	return m, m.inputs.Update(msg)
}

func (m installModel) View() string {
	var b strings.Builder
	// TODO This should be instantiated by the command
	desc, _ := style.Glamour.Render(`## Welcome!

This is the Flightcrew installation CLI! To get started, please fill in the information below.`)

	b.WriteString(desc)
	b.WriteString(m.inputs.View())
	b.WriteString("\n\n")

	if m.confirming {
		if !m.hasErrors {
			b.WriteString(m.confirmYesButton.View(m.index == m.inputs.Len()))
			b.WriteRune(' ')
		}
		b.WriteString(m.confirmNoButton.View(m.index == m.inputs.Len()+1))
	} else {
		b.WriteString(m.submitButton.View(m.index == m.inputs.Len()))
	}

	b.WriteString("\n\n")

	b.WriteString(style.Required("*"))
	b.WriteString(" - required\n\n")
	b.WriteString(style.Help("ctrl+c/esc: quit • ←/→/↑/↓: nav • enter: proceed"))

	if len(m.logStatements) > 0 {
		b.WriteRune('\n')
		for _, str := range m.logStatements {
			b.WriteString(str)
			b.WriteRune('\n')
		}
	}

	return b.String()
	// return wordwrap.String(b.String(), m.width)
}
