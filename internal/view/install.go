package view

import (
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/gcp"
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
	keyIAMFile           = "${IAM_FILE}"
	keyPermissions       = "${PERMISSIONS}"
	keyPlatform          = "${PLATFORM}"
)

type inputEntry struct {
	Title    string
	HelpText string
	Required bool
	Default  string
	// If the value is to be converted, this is only valid when model.confirming is true.
	Converted string
	Error     string
	Freeform  textinput.Model
	Selector  *HorizontalSelector
}

type installModel struct {
	params InstallParams

	width  int
	height int

	inputs     map[string]*inputEntry
	inputKeys  []string
	focusIndex int

	titleStyle lipgloss.Style

	confirming       bool
	submitButton     *Button
	confirmYesButton *Button
	confirmNoButton  *Button

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

	m := installModel{
		params: params,
		inputs: make(map[string]*inputEntry),
		inputKeys: []string{
			keyProject,
			keyVirtualMachine,
			keyAPIToken,
			keyZone,
			keyTowerVersion,
			keyIAMServiceAccount,
			keyIAMRole,
			keyIAMFile,
			keyPlatform,
			keyPermissions,
		},
		logStatements: make([]string, 0),
	}

	m.submitButton, _ = NewButton("Submit", 12)
	m.confirmYesButton, _ = NewButton("Continue", 12)
	m.confirmNoButton, _ = NewButton("Edit", 12)

	var maxTitleLength int

	for _, key := range m.inputKeys {
		input := &inputEntry{
			Freeform: textinput.New(),
		}
		input.Freeform.CursorStyle = cursorStyle
		input.Freeform.CharLimit = 32

		switch key {
		case keyProject:
			if len(params.ProjectName) > 0 {
				input.Freeform.SetValue(params.ProjectName)
				input.Default = params.ProjectName
			} else {
				input.Freeform.Placeholder = "project-id-1234"
			}
			input.Freeform.Focus()
			input.Freeform.PromptStyle = style.Focused
			input.Freeform.TextStyle = style.Focused
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
			input.Required = true

		case keyVirtualMachine:
			if params.VirtualMachineName != "" {
				input.Freeform.SetValue(params.VirtualMachineName)
			}
			input.Freeform.Placeholder = "flightcrew-control-tower"
			input.Freeform.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."

		case keyZone:
			if params.Zone != "" {
				input.Freeform.SetValue(params.Zone)
			}
			input.Freeform.Placeholder = "us-central"
			input.Freeform.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."

		case keyTowerVersion:
			if params.TowerVersion != "" {
				input.Freeform.SetValue(params.TowerVersion)
			}
			input.Freeform.Placeholder = "stable"
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"

		case keyAPIToken:
			if len(params.Token) > 0 {
				input.Freeform.SetValue(params.Token)
			}
			input.Freeform.Placeholder = "api-token"
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."

		case keyIAMServiceAccount:
			input.Title = "IAM Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.Freeform.SetValue("flightcrew-runner-test-chris")
			input.HelpText = "IAM Service Account is the name of the (to be created) service account to run the Flightcrew Tower."

		case keyIAMRole:
			input.Title = "IAM Role"
			input.Default = "flightcrew.gae.read.only"
			input.Freeform.SetValue("flightcrew.gae.read.only")
			input.HelpText = "IAM Role is the name of the (to be created) IAM role defining permissions to run the Flightcrew Tower."

		case keyIAMFile:
			input.Title = "IAM File"
			input.Default = "flightcrew.gae.read.only"
			input.Freeform.SetValue("gcp/gae/iam_readonly.yaml")
			input.HelpText = "IAM File provides the list of permissions to attach to the (to be created) IAM Role."

		case keyPlatform:
			input.Title = "Platform"
			input.HelpText = "Platform is which Google Cloud Provider resources Flightcrew will read in."
			input.Selector = NewHorizontalSelector([]string{"App Engine", "Compute Engine"})

		case keyPermissions:
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			input.Selector = NewHorizontalSelector([]string{"Read", "Write"})
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
			switch s {
			case "enter":
				if m.focusIndex == len(m.inputs) {
					if !m.confirming {
						m.confirming = true
						m.convertValues()
						return m, nil
					}

					args := make(map[string]string)
					for key, wInput := range m.inputs {
						if len(wInput.Converted) > 0 {
							args[key] = wInput.Converted
						} else if wInput.Selector != nil {
							args[key] = wInput.Selector.Value()
						} else {
							args[key] = wInput.Freeform.Value()
						}
					}

					return NewRunModel(args), nil
				}

				if m.focusIndex == len(m.inputs)+1 {
					m.confirming = false
					m.focusIndex = len(m.inputs)
					m.resetConverted()
					return m, nil
				}

				m.focusIndex++
				for ; m.focusIndex < len(m.inputs); m.focusIndex++ {
					input := m.getInput(m.focusIndex)
					if input.Selector != nil {
						if len(input.Selector.Value()) == 0 {
							break
						}
					} else if len(input.Freeform.Value()) == 0 {
						break
					}
				}

			case "up", "shift+tab":
				if !m.confirming {
					m.focusIndex--
				}

			case "down":
				if !m.confirming {
					m.focusIndex++
				}

			case "tab":
				if m.confirming {
					if m.focusIndex == len(m.inputs)+1 {
						m.focusIndex = len(m.inputs)
					} else if m.focusIndex == len(m.inputs) {
						m.focusIndex = len(m.inputs) + 1
					}
				} else {
					m.focusIndex++
				}

			case "left":
				if m.confirming {
					m.focusIndex--
				} else {
					input := m.getInput(m.focusIndex)
					if input.Selector != nil {
						input.Selector.MoveLeft()
					}
				}

			case "right":
				if m.confirming {
					m.focusIndex++
				} else {
					input := m.getInput(m.focusIndex)
					if input.Selector != nil {
						input.Selector.MoveRight()
					}
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

			return m, m.updateFocus()
		}
	}

	// Handle character input and blinking
	return m, m.updateInputs(msg)
}

func (m *installModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i < len(m.inputs); i++ {
		input := m.getInput(i)
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = input.Freeform.Focus()
			input.Freeform.PromptStyle = style.Focused
			input.Freeform.TextStyle = style.Focused
			continue
		}
		// Remove focused state
		input.Freeform.Blur()
		input.Freeform.PromptStyle = style.None
		input.Freeform.TextStyle = style.None
	}
	return tea.Batch(cmds...)
}

func (m *installModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	var i int

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for k, val := range m.inputs {
		if m.inputs[k].Selector != nil {
			m.inputs[k].Freeform, cmds[i] = val.Freeform.Update(msg)
		}
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
		input := m.getInput(i)
		b.WriteString(m.titleStyle.Render(input.Title))
		if input.Required {
			b.WriteString(style.Required("*"))
		} else {
			b.WriteRune(' ')
		}
		b.WriteString(": ")

		if m.confirming {
			if input.Selector != nil {
				b.WriteString(input.Selector.Value())
			} else {
				b.WriteString(input.Freeform.Value())
			}
			if len(input.Converted) > 0 {
				b.WriteString(" → ")
				b.WriteString(input.Converted)
			}

			if len(input.Error) > 0 {
				b.WriteString(" ❗️ ")
				b.WriteString(input.Error)
			}
		} else {
			if input.Selector != nil {
				b.WriteString("> ")
				b.WriteString(input.Selector.View())
			} else {
				b.WriteString(input.Freeform.View())
			}
		}

		b.WriteRune('\n')
	}

	if m.focusIndex < len(m.inputKeys) {
		b.WriteString(m.getInput(m.focusIndex).HelpText)
		b.WriteRune('\n')
	} else if !m.confirming {
		b.WriteString(m.defaultHelpText)
		b.WriteRune('\n')
	}

	b.WriteString("\n\n")

	if m.confirming {
		b.WriteString(m.confirmYesButton.View(m.focusIndex == len(m.inputs)))
		b.WriteRune(' ')
		b.WriteString(m.confirmNoButton.View(m.focusIndex == len(m.inputs)+1))
	} else {
		b.WriteString(m.submitButton.View(m.focusIndex == len(m.inputs)))
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

	return wordwrap.String(b.String(), m.width)
}

func (m *installModel) getInput(i int) *inputEntry {
	k := m.inputKeys[i]
	return m.inputs[k]
}

func (m *installModel) convertValues() {
	for k, val := range m.inputs {
		switch k {
		case keyTowerVersion:
			version, err := gcp.GetTowerImageVersion(val.Freeform.Value())
			if err != nil {
				val.Error = err.Error()
				debug.Output("convert tower version got error: %v", err)
				continue
			}

			val.Converted = version
			debug.Output("convert tower version is %s", version)
		}
	}
}

func (m *installModel) resetConverted() {
	for _, val := range m.inputs {
		val.Converted = ""
		val.Error = ""
	}
}
