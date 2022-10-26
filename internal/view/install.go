package view

import (
	"strings"

	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	keyProject           = "${GOOGLE_PROJECT_ID}"
	keyTowerVersion      = "${TOWER_VERSION}"
	keyZone              = "${ZONE}"
	keyVirtualMachine    = "${VIRTUAL_MACHINE}"
	keyAPIToken          = "${API_TOKEN}"
	keyIAMRole           = "${IAM_ROLE}"
	keyIAMServiceAccount = "${SERVICE_ACCOUNT}"
)

type installModel struct {
	params InstallParams

	inputs     map[string]*wrappedInput
	inputKeys  []string
	focusIndex int

	titleStyle lipgloss.Style

	confirming bool

	cursorMode textinput.CursorMode
}

type InstallParams struct {
	VirtualMachineName string
	ProjectName        string
	Zone               string
	TowerVersion       string
	Token              string
	IAMRole            string
	IAMFile            string
	ServiceAccount     string
}

func NewInstallModel(params InstallParams) installModel {
	if params.VirtualMachineName == "" {
		params.VirtualMachineName = "flightcrew-control-tower"
	}

	if params.Zone == "" {
		params.Zone = "us-central"
	}

	if params.TowerVersion == "" {
		params.TowerVersion = "latest"
	}

	const defaultWidth = 20

	m := installModel{
		params: params,
		inputs: make(map[string]*wrappedInput),
		inputKeys: []string{
			keyProject,
			keyVirtualMachine,
			keyAPIToken,
			keyZone,
			keyTowerVersion,
			keyIAMServiceAccount,
			keyIAMRole,
		},
	}

	var maxTitleLength int

	for _, key := range m.inputKeys {
		input := &wrappedInput{
			Input: textinput.New(),
		}
		input.Input.CursorStyle = cursorStyle
		input.Input.CharLimit = 32

		switch key {
		case keyProject:
			if params.ProjectName != "" {
				input.Input.SetValue(params.ProjectName)
				input.Default = params.ProjectName
			} else {
				input.Input.Placeholder = "project-id-1234"
			}
			input.Input.Focus()
			input.Input.PromptStyle = style.Focused
			input.Input.TextStyle = style.Focused
			input.Title = "Project ID"
			input.Required = true

		case keyVirtualMachine:
			if params.VirtualMachineName != "" {
				input.Input.SetValue(params.VirtualMachineName)
			}
			input.Input.Placeholder = "flightcrew-control-tower"
			input.Input.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true

		case keyZone:
			if params.Zone != "" {
				input.Input.SetValue(params.Zone)
			}
			input.Input.Placeholder = "us-central"
			input.Input.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"

		case keyTowerVersion:
			if params.TowerVersion != "" {
				input.Input.SetValue(params.TowerVersion)
			}
			input.Input.Placeholder = "latest"
			input.Title = "Version"
			input.Default = "latest"

		case keyAPIToken:
			input.Input.Placeholder = "api-token"
			input.Title = "API Token"
			//			input.HelpText = "This is the API token provided by Flightcrew to identify your organization."

		case keyIAMServiceAccount:
			input.Title = "IAM Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.Input.SetValue("flightcrew-runner-test-chris")

		case keyIAMRole:
			input.Title = "IAM Role"
			input.Default = "flightcrew.gae.read.only"
			input.Input.SetValue("flightcrew.gae.read.only")

		}

		m.inputs[key] = input

		if titleLength := len(m.inputs[key].Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	m.titleStyle = lipgloss.NewStyle().Align(lipgloss.Right).Width(maxTitleLength)

	return m
}

func (m installModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode = textinput.CursorBlink
			cmds := make([]tea.Cmd, len(m.inputKeys))
			for i, key := range m.inputKeys {
				cmds[i] = m.inputs[key].Input.SetCursorMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" && m.focusIndex == len(m.inputs) {
				if !m.confirming {
					m.confirming = true
					return m, nil
				}

				args := make(map[string]string)
				for key, wInput := range m.inputs {
					args[key] = wInput.Input.Value()
				}

				return NewRunModel(args), nil
			} else if s == "enter" && m.focusIndex == len(m.inputs)+1 {
				m.confirming = false
				return m, nil
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				if !m.confirming || m.focusIndex > len(m.inputs)+1 {
					m.focusIndex = 0
				}
			} else if m.focusIndex < 0 {
				if m.confirming {
					m.focusIndex = len(m.inputs) + 1

				} else {
					m.focusIndex = len(m.inputs)
				}
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				key := m.inputKeys[i]
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[key].Input.Focus()
					m.inputs[key].Input.PromptStyle = style.Focused
					m.inputs[key].Input.TextStyle = style.Focused
					continue
				}
				// Remove focused state
				m.inputs[key].Input.Blur()
				m.inputs[key].Input.PromptStyle = style.None
				m.inputs[key].Input.TextStyle = style.None
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *installModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	var i int

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for k, val := range m.inputs {
		m.inputs[k].Input, cmds[i] = val.Input.Update(msg)
		i++
	}

	return tea.Batch(cmds...)
}

func (m installModel) View() string {
	var b strings.Builder
	for i := range m.inputKeys {
		k := m.inputKeys[i]
		b.WriteString(m.titleStyle.Render(m.inputs[k].Title))
		if m.inputs[k].Required {
			b.WriteString(requiredStyle.Render("*"))
		} else {
			b.WriteRune(' ')
		}
		b.WriteString(": ")

		if m.confirming {
			b.WriteString(m.inputs[k].Input.Value())
		} else {
			b.WriteString(m.inputs[k].Input.View())
		}

		if len(m.inputs[k].HelpText) > 0 {
			b.WriteRune('\n')
			b.WriteString(helpStyle.Render(m.inputs[k].HelpText))
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	b.WriteString("\n\n")

	if m.confirming {
		if m.focusIndex == len(m.inputs) {
			b.WriteString(style.Focused.Render(confirmYesText))
		} else {
			b.WriteString(style.Blurred.Render(confirmYesText))
		}

		b.WriteRune(' ')

		if m.focusIndex == len(m.inputs)+1 {
			b.WriteString(style.Focused.Render(confirmNoText))
		} else {
			b.WriteString(style.Blurred.Render(confirmNoText))
		}
	} else {
		if m.focusIndex == len(m.inputs) {
			b.WriteString(style.Focused.Render(submitText))
		} else {
			b.WriteString(style.Blurred.Render(submitText))
		}
	}

	b.WriteString("\n\n")

	b.WriteString(requiredStyle.Render("*"))
	b.WriteString(" - required\n\n")
	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
