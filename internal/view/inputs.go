package view

import (
	"strings"

	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Inputs is a variable number of inputs that can be changed. This interface allows
// pre-defined defaults that can be modified according to cloud providers' specific
// requirements.
type Inputs interface {
	// Len returns the number of inputs in this grouping. The caller will assume that there are
	// Len() number of inputs, so any index outside of [0, len) will be used for keeping state
	// in the caller.
	Len() int

	// Validate gives the library a chance to convert any values that need to be converted
	// or make sure that all inputs have valid values.
	Validate() bool
	// Reset goes from Validation state to Edit state.
	Reset()

	View() string
	Update(msg tea.Msg) tea.Cmd

	// NextEmpty should return the index of the next empty user input field.
	NextEmpty(i int) int
	// Focus will be called every time the index changes. The index will assume that
	// [0, inputs.Len()) refers to the inputs within this implementation. The implementation
	// should be resilient to array out of bound.
	Focus(i int) tea.Cmd

	// Args should return the a map from key (managed by the implementer) to value for
	// variables that should be replaced in the Commands below. Take care to create
	// keys that will not be prefixes to commonly occurring words.
	// e.g. Defining $PRE and $PREFIX may cause unexpected results.
	Args() map[string]string
	// Commands should return the list of commands that should be run after the inputs
	// have been confirmed. Commands and descriptions will be updated with the input
	// values based on the keys that are provided by the implementation. The run model will
	// use and modify the same slice, so the implementation will be able to share the state.
	Commands() []*command.Model
}

type inputsModel struct {
	width  int
	height int

	inputs     Inputs
	index      int
	hasErrors  bool
	confirming bool

	submitButton     *button.Button
	confirmYesButton *button.Button
	confirmNoButton  *button.Button

	logStatements []string
}

func NewInputsModel(inputs Inputs) inputsModel {
	m := inputsModel{
		inputs:        inputs,
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

func (m inputsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

					return NewRunModel(m.inputs), nil
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

func (m inputsModel) View() string {
	var b strings.Builder
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

	b.WriteString(style.Help("ctrl+c/esc: quit • ←/→/↑/↓: nav • enter: proceed"))

	if len(m.logStatements) > 0 {
		b.WriteRune('\n')
		for _, str := range m.logStatements {
			b.WriteString(str)
			b.WriteRune('\n')
		}
	}

	return b.String()
}
