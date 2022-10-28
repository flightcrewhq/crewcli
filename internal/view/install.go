package view

import (
	"fmt"
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/gcp"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
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
	keyIAMFile           = "${IAM_FILE}"
	keyPermissions       = "${PERMISSIONS}"
	keyRPCHost           = "${RPC_HOST}"
	keyPlatform          = "${PLATFORM}"
	keyTrafficRouter     = "${TRAFFIC_ROUTER}"
)

var (
	filenameReplacer = strings.NewReplacer(
		".", "_",
		"/", "_",
		":", "_",
		" ", "_",
	)
)

type inputEntry struct {
	Title    string
	HelpText string
	Required bool
	Default  string
	// If the value is to be converted, this is only valid when model.confirming is true.
	Converted string
	Message   string
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
	hasErrors  bool

	titleStyle lipgloss.Style

	confirming       bool
	submitButton     *button.Button
	confirmYesButton *button.Button
	confirmNoButton  *button.Button

	defaultHelpText string
	tempDir         string
	args            map[string]string

	logStatements []string
}

type InstallParams struct {
	VirtualMachineName  string
	ProjectName         string
	Zone                string
	TowerVersion        string
	Token               string
	IAMRole             string
	IAMFile             string
	ServiceAccount      string
	PlatformDisplayName string
	ReadOnly            bool
}

func NewInstallModel(params InstallParams, tempDir string) installModel {
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
			keyPlatform,
			keyPermissions,
			keyZone,
			keyTowerVersion,
			keyIAMServiceAccount,
		},
		logStatements: make([]string, 0),
	}

	m.submitButton, _ = button.New("Submit", 12)
	m.confirmYesButton, _ = button.New("Continue", 12)
	m.confirmNoButton, _ = button.New("Edit", 12)

	m.tempDir = tempDir
	m.args = make(map[string]string)
	m.args[keyRPCHost] = "api.flightcrew.io"
	m.args[keyTrafficRouter] = ""

	var maxTitleLength int

	for _, key := range m.inputKeys {
		input := &inputEntry{
			Freeform: textinput.New(),
		}
		input.Freeform.CursorStyle = style.Focused.Copy()
		input.Freeform.CharLimit = 32

		switch key {
		case keyProject:
			input.Freeform.CharLimit = 0
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
			input.Title = "Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.Freeform.SetValue("flightcrew-runner-test-chris")
			input.HelpText = "Service Account is the name of the (to be created) IAM service account to run the Flightcrew Tower."

		case keyPlatform:
			input.Title = "Platform"
			input.HelpText = "Platform is which Google Cloud Provider resources Flightcrew will read in."
			input.Selector = NewHorizontalSelector([]string{"App Engine", "Compute Engine"})
			if len(params.PlatformDisplayName) > 0 {
				input.Selector.SetValue(params.PlatformDisplayName)
			}

		case keyPermissions:
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			input.Selector = NewHorizontalSelector([]string{constants.Read, constants.Write})
			if params.ReadOnly {
				input.Selector.SetValue(constants.Read)
			} else {
				input.Selector.SetValue(constants.Write)
			}

		}

		m.inputs[key] = input

		if titleLength := len(m.inputs[key].Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	// Format help text.
	wrappedText, _ := style.Glamour.Render("> Edit a particular entry to see help text here.\n> Otherwise, press enter to proceed.")
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

	m.titleStyle = lipgloss.NewStyle().Align(lipgloss.Right).Width(maxTitleLength).MarginLeft(2)
	m.nextEmptyInput()
	m.updateFocus()

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
				if m.focusIndex == len(m.inputs) {
					if !m.confirming {
						m.confirming = true
						m.convertValues()
						if m.hasErrors {
							m.focusIndex = len(m.inputs) + 1
						}
						return m, nil
					}

					for key, wInput := range m.inputs {
						if len(wInput.Converted) > 0 {
							m.args[key] = wInput.Converted
						} else if wInput.Selector != nil {
							m.args[key] = wInput.Selector.Value()
						} else {
							m.args[key] = wInput.Freeform.Value()
						}
					}

					return NewRunModel(m.args), nil
				}

				if m.focusIndex == len(m.inputs)+1 {
					m.confirming = false
					m.hasErrors = false
					m.focusIndex = len(m.inputs)
					m.resetConverted()
					return m, nil
				}

				m.focusIndex++
				m.nextEmptyInput()

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
					if !m.hasErrors {
						m.focusIndex--
					}
				} else if m.focusIndex < len(m.inputs) {
					input := m.getInput(m.focusIndex)
					if input.Selector != nil {
						input.Selector.MoveLeft()
					}
					cmds = append(cmds, m.updateInputs(msg))
				}

			case "right":
				if m.confirming {
					m.focusIndex++
				} else if m.focusIndex < len(m.inputs) {
					input := m.getInput(m.focusIndex)
					if input.Selector != nil {
						input.Selector.MoveRight()
					}
					cmds = append(cmds, m.updateInputs(msg))
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

			cmds = append(cmds, m.updateFocus())
			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	return m, m.updateInputs(msg)
}

func (m *installModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.inputs))
	for i := 0; i < len(m.inputs); i++ {
		input := m.getInput(i)
		if i == m.focusIndex {
			// Set focused state
			cmds = append(cmds, input.Freeform.Focus())
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
		if m.inputs[k].Selector == nil {
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

			if len(input.Error) > 0 {
				b.WriteString(" ❗️ ")
				b.WriteString(style.Error(input.Error))
			} else if len(input.Message) > 0 {
				b.WriteString(" → ")
				b.WriteString(style.Convert(input.Message))
			} else if len(input.Converted) > 0 {
				b.WriteString(" → ")
				b.WriteString(style.Convert(input.Converted))
			}

		} else {
			if input.Selector != nil {
				b.WriteString(input.Selector.View(m.focusIndex == i))
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
		if !m.hasErrors {
			b.WriteString(m.confirmYesButton.View(m.focusIndex == len(m.inputs)))
			b.WriteRune(' ')
		}
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

	return b.String()
	// return wordwrap.String(b.String(), m.width)
}

func (m *installModel) getInput(i int) *inputEntry {
	k := m.inputKeys[i]
	return m.inputs[k]
}

func (m *installModel) convertValues() {
	for k, val := range m.inputs {
		if val.Required {
			if val.Selector != nil &&
				len(val.Selector.Value()) == 0 {
				val.Error = "required"
				m.hasErrors = true
				continue
			}

			if val.Selector == nil &&
				len(val.Freeform.Value()) == 0 {
				val.Error = "required"
				m.hasErrors = true
				continue
			}
		}

		switch k {
		case keyTowerVersion:
			version, err := gcp.GetTowerImageVersion(val.Freeform.Value())
			if err != nil {
				val.Error = err.Error()
				debug.Output("convert tower version got error: %v", err)
				break
			}

			val.Converted = version
			debug.Output("convert tower version is %s", version)

		case keyPlatform:
			displayName := val.Selector.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				val.Error = "invalid platform"
				break
			}

			val.Converted = platform

		case keyPermissions:
			platformInput := m.inputs[keyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Selector.Value()]
			if !ok {
				break
			}

			perms, ok := constants.PlatformPermissions[platform]
			if !ok {
				val.Error = "platform has no permissions"
				break
			}

			permission := val.Selector.Value()
			permSettings, ok := perms[permission]
			if !ok {
				val.Error = fmt.Sprintf("%s permissions are not supported for platform '%s'", permission, platformInput.Selector.Value())
				break
			}

			f, err := os.CreateTemp(m.tempDir, filenameReplacer.Replace(fmt.Sprintf("%s_%s", permission, platform)))
			if err != nil {
				val.Error = "create temp file to put permissions YAML"
				break
			}

			if _, err := f.WriteString(permSettings.Content); err != nil {
				val.Error = err.Error()
				break
			}

			if err := f.Close(); err != nil {
				val.Error = err.Error()
				break
			}

			if permission == constants.Write {
				m.args[keyTrafficRouter] = fmt.Sprintf(`
  --container-env=TRAFFIC_ROUTER=%s \`, platform)
			}

			m.args[keyIAMFile] = f.Name()
			m.args[keyIAMRole] = permSettings.Role
			val.Message = fmt.Sprintf("see %s", f.Name())
		}

		if len(val.Error) > 0 {
			m.hasErrors = true
		}
	}
}

func (m *installModel) resetConverted() {
	for _, val := range m.inputs {
		val.Converted = ""
		val.Error = ""
		val.Message = ""
	}
}

func (m *installModel) nextEmptyInput() {
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
}
