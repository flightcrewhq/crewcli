package view

import (
	"strings"

	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/wrapinput"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputsModel struct {
	submitButton     *button.Button
	confirmYesButton *button.Button
	confirmNoButton  *button.Button

	requiredHelpText string
	defaultHelpText  string
	description      string

	controller controller.Inputs
	inputs     []*wrapinput.Model
	index      int

	width  int
	height int

	hasErrors  bool
	confirming bool
}

func NewInputsModel(controller controller.Inputs) InputsModel {
	m := InputsModel{
		controller: controller,
	}

	m.submitButton, _ = button.New("Submit", 12)
	m.confirmYesButton, _ = button.New("Continue", 12)
	m.confirmNoButton, _ = button.New("Edit", 12)

	allInputs := controller.GetAllInputs()
	m.updateTitleStyles(allInputs)
	m.updateHelpText(allInputs)

	m.description, _ = style.Glamour.Render(strings.Replace(`## Welcome!

This is the Flightcrew ${NAME} CLI! To get started, please fill in the information below.`, "${NAME}", controller.GetName(), 1))

	m.inputs = controller.GetInputs()
	m.updateInput(nil)
	m.moveToNextEmptyInput()
	m.updateFocusFrom(0)

	return m
}

func (m *InputsModel) updateHelpText(inputs []*wrapinput.Model) {
	countTrimmedRenderNewlines := func(text string) (string, int) {
		wrappedText, _ := style.Glamour.Render(controller.DefaultHelpText)
		trimmed := strings.Trim(wrappedText, "\n")
		return trimmed, strings.Count(trimmed, "\n")
	}

	// Format help text.
	var defaultLines int
	m.defaultHelpText, defaultLines = countTrimmedRenderNewlines(controller.DefaultHelpText)

	maxLines := defaultLines
	lineCounts := make([]int, len(inputs))
	for i := range inputs {
		inputs[i].HelpText, lineCounts[i] = countTrimmedRenderNewlines(inputs[i].HelpText)
		if lineCounts[i] > maxLines {
			maxLines = lineCounts[i]
		}
	}

	adjustNewlines := func(text string, currCount, desiredCount int) string {
		if diff := desiredCount - currCount; diff > 0 {
			return text + strings.Repeat("\n", diff)
		}
		return text
	}
	m.defaultHelpText = adjustNewlines(m.defaultHelpText, defaultLines, maxLines)
	for i := range inputs {
		inputs[i].HelpText = adjustNewlines(inputs[i].HelpText, lineCounts[i], maxLines)
	}
}

func (m *InputsModel) updateTitleStyles(inputs []*wrapinput.Model) {
	var maxTitleLength int
	for _, input := range inputs {
		if titleLength := len(input.Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	// Format titles
	titleStyle := lipgloss.NewStyle().Align(lipgloss.Right).Width(maxTitleLength).MarginLeft(2).Render
	for i := range inputs {
		inputs[i].Title = titleStyle(inputs[i].Title)
	}

	m.requiredHelpText = titleStyle(style.Required("*")) + style.HelpColor.Render(" - required")
}

func (m InputsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height // height is available too

	case tea.KeyMsg:
		cmds := make([]tea.Cmd, 0)
		oldIndex := m.index
		switch s := msg.String(); s {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down", "left", "right":
			switch s {
			case "enter":
				if m.index == len(m.inputs) {
					if !m.confirming {
						m.confirming = true
						m.hasErrors = !m.controller.Validate(m.inputs)
						if m.hasErrors {
							m.index = len(m.inputs) + 1
						}
						return m, nil
					}

					return NewRunModel(m.controller.GetRunController()), nil
				}

				if m.index == len(m.inputs)+1 {
					m.confirming = false
					m.hasErrors = false
					m.index = len(m.inputs)
					m.controller.Reset(m.inputs)
					return m, nil
				}

				m.moveToNextEmptyInput()

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
					if m.index == len(m.inputs)+1 {
						m.index = len(m.inputs)
					} else if m.index == len(m.inputs) {
						m.index = len(m.inputs) + 1
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

			if m.index > len(m.inputs) {
				if m.confirming && m.index > len(m.inputs)+1 {
					m.index = len(m.inputs) + 1
				} else if !m.confirming {
					m.index = 0
				}
			} else if m.confirming && m.index < len(m.inputs) {
				m.index = len(m.inputs)
			} else if !m.confirming && m.index < 0 {
				m.index = len(m.inputs)
			}

			if m.confirming {
				break
			}

			cmds = append(cmds, m.updateFocusFrom(oldIndex), m.updateInput(msg))
			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	return m, m.updateInput(msg)
}

func (m *InputsModel) updateFocusFrom(oldIndex int) tea.Cmd {
	if oldIndex == m.index {
		return nil
	}

	if oldIndex < len(m.inputs) {
		m.inputs[oldIndex].Blur()
	}

	if m.index < len(m.inputs) {
		return m.inputs[m.index].Focus()
	}

	return nil
}

func (m *InputsModel) updateInput(msg tea.Msg) tea.Cmd {
	if m.index >= len(m.inputs) {
		return nil
	}

	var cmd tea.Cmd
	*m.inputs[m.index], cmd = m.inputs[m.index].Update(msg)
	m.inputs = m.controller.GetInputs()
	return cmd
}

func (m *InputsModel) moveToNextEmptyInput() {
	for ; m.index < len(m.inputs); m.index++ {
		if len(m.inputs[m.index].Value()) == 0 {
			break
		}
	}
}

func (m InputsModel) View() string {
	var b strings.Builder
	b.WriteString(m.description)
	b.WriteRune('\n')
	b.WriteString(m.viewInputs())
	b.WriteString("\n\n")

	if m.confirming {
		if !m.hasErrors {
			b.WriteString(m.confirmYesButton.View(m.index == len(m.inputs)))
			b.WriteRune(' ')
		}
		b.WriteString(m.confirmNoButton.View(m.index == len(m.inputs)+1))
	} else {
		b.WriteString(m.submitButton.View(m.index == len(m.inputs)))
	}

	b.WriteString("\n\n")

	b.WriteString(style.Help("ctrl+c/esc: quit • ←/→/↑/↓: nav • enter: proceed"))
	return b.String()
}

func (m InputsModel) viewInputs() string {
	var b strings.Builder
	for _, input := range m.inputs {
		b.WriteString(input.View(wrapinput.ViewParams{
			ShowValue: m.confirming,
		}))
		b.WriteRune('\n')
	}

	if !m.confirming {
		b.WriteString(m.requiredHelpText)
		b.WriteRune('\n')
	}

	if m.index < len(m.inputs) {
		b.WriteString(m.inputs[m.index].HelpText)
		b.WriteRune('\n')
	} else if !m.confirming {
		b.WriteString(m.defaultHelpText)
		b.WriteRune('\n')
	}

	return b.String()
}
