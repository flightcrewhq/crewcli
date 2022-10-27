package view

import (
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
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

	width  int
	height int

	inputs     map[string]*wrappedInput
	inputKeys  []string
	focusIndex int

	titleStyle lipgloss.Style

	confirming bool

	defaultHelpText string

	logStatements []string
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
	debug.Output("new install model!")
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
		logStatements: make([]string, 0),
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
			if len(params.ProjectName) > 0 {
				input.Input.SetValue(params.ProjectName)
				input.Default = params.ProjectName
			} else {
				input.Input.Placeholder = "project-id-1234"
			}
			input.Input.Focus()
			input.Input.PromptStyle = style.Focused
			input.Input.TextStyle = style.Focused
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
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
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."

		case keyZone:
			if params.Zone != "" {
				input.Input.SetValue(params.Zone)
			}
			input.Input.Placeholder = "us-central"
			input.Input.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."

		case keyTowerVersion:
			if params.TowerVersion != "" {
				input.Input.SetValue(params.TowerVersion)
			}
			input.Input.Placeholder = "latest"
			input.Title = "Tower Version"
			input.Default = "latest"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"

		case keyAPIToken:
			if len(params.Token) > 0 {
				input.Input.SetValue(params.Token)
			}
			input.Input.Placeholder = "api-token"
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."

		case keyIAMServiceAccount:
			input.Title = "IAM Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.Input.SetValue("flightcrew-runner-test-chris")
			input.HelpText = "IAM Service Account is the name of the (to be created) service account to run the Flightcrew Tower."

		case keyIAMRole:
			input.Title = "IAM Role"
			input.Default = "flightcrew.gae.read.only"
			input.Input.SetValue("flightcrew.gae.read.only")
			input.HelpText = "IAM Role is the name of the (to be created) IAM role defining permissions to run the Flightcrew Tower."

		}

		m.inputs[key] = input

		if titleLength := len(m.inputs[key].Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	// Format help text.
	wrappedText, _ := style.Glamour.Render("> Edit a particular entry to see help text here.")
	m.defaultHelpText = strings.Trim(wrappedText, "\n")
	defaultLines := strings.Count(m.defaultHelpText, "\n")
	maxLines := defaultLines
	lineCounts := make(map[string]int)
	for k := range m.inputs {
		wrappedText, _ := style.Glamour.Render("> " + m.inputs[k].HelpText)
		m.inputs[k].HelpText = strings.Trim(wrappedText, "\n")
		lineCounts[k] = strings.Count(m.inputs[k].HelpText, "\n")

		if lineCounts[k] > maxLines {
			maxLines = lineCounts[k]
		}
	}
	if defaultDiff := maxLines - defaultLines; defaultDiff > 0 {
		m.defaultHelpText += strings.Repeat("\n", defaultDiff)
	}
	for k := range m.inputs {
		if diff := maxLines - lineCounts[k]; diff > 0 {
			m.inputs[k].HelpText += strings.Repeat("\n", diff)
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height // height is available too

	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down", "left", "right":
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
				m.focusIndex = len(m.inputs)
				return m, nil
			}

			// Cycle indexes
			if !m.confirming {
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else if s == "enter" {
					m.focusIndex++
					for ; m.focusIndex < len(m.inputs); m.focusIndex++ {
						if len(m.inputs[m.inputKeys[m.focusIndex]].Input.Value()) == 0 {
							break
						}
					}
				} else {
					m.focusIndex++
				}
			} else {
				if s == "left" {
					m.focusIndex--
				} else if s == "right" {
					m.focusIndex++
				}
			}

			if m.focusIndex > len(m.inputs) {
				if m.confirming && m.focusIndex > len(m.inputs)+1 {
					m.focusIndex = len(m.inputs) + 1
				} else if !m.confirming {
					m.focusIndex = 0
				}
			} else if m.confirming && m.focusIndex < len(m.inputs) {
				m.focusIndex = len(m.inputs)
			} else if !m.confirming && m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			if m.confirming {
				break
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
	desc, _ := style.Glamour.Render(`## Welcome!

This is the Flightcrew installation CLI! To get started, please fill in the information below.`)

	b.WriteString(desc)

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

		b.WriteRune('\n')
	}

	if m.focusIndex < len(m.inputKeys) {
		b.WriteString(m.inputs[m.inputKeys[m.focusIndex]].HelpText)
		b.WriteRune('\n')
	} else if !m.confirming {
		b.WriteString(m.defaultHelpText)
		b.WriteRune('\n')
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
	b.WriteString(style.Help("ctrl+c/esc: quit • ←/→/↑/↓: nav • enter: proceed"))

	if len(m.logStatements) > 0 {
		b.WriteRune('\n')
		for _, str := range m.logStatements {
			b.WriteString(str)
			b.WriteRune('\n')
		}
	}

	return wordwrap.String(b.String(), m.width)
}
