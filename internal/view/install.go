package view

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/gcp"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/wrapinput"
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
	keyImagePath         = "${IMAGE_PATH}"
)

var (
	filenameReplacer = strings.NewReplacer(
		".", "_",
		"/", "_",
		":", "_",
		" ", "_",
	)
)

type installModel struct {
	width  int
	height int

	// TODO: Move the GCP stuff into a separate place. Probably group this entire block and turn it into its own model?
	params     gcp.InstallParams
	inputs     map[string]*wrapinput.Model
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

func NewInstallModel(params gcp.InstallParams, tempDir string) installModel {
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
		inputs: make(map[string]*wrapinput.Model),
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
		var input wrapinput.Model

		switch key {
		case keyProject:
			input = wrapinput.NewFreeForm()
			input.Freeform.CharLimit = 0
			if len(params.ProjectName) > 0 {
				input.SetValue(params.ProjectName)
				input.Default = params.ProjectName
			} else {
				input.Freeform.Placeholder = "project-id-1234"
			}
			_ = input.Focus()
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
			input.Required = true

		case keyVirtualMachine:
			input = wrapinput.NewFreeForm()
			if params.VirtualMachineName != "" {
				input.SetValue(params.VirtualMachineName)
			}
			input.Freeform.Placeholder = "flightcrew-control-tower"
			input.Freeform.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."

		case keyZone:
			input = wrapinput.NewFreeForm()
			if params.Zone != "" {
				input.SetValue(params.Zone)
			}
			input.Freeform.Placeholder = "us-central"
			input.Freeform.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."

		case keyTowerVersion:
			input = wrapinput.NewFreeForm()
			if params.TowerVersion != "" {
				input.SetValue(params.TowerVersion)
			}
			input.Freeform.Placeholder = "stable"
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"

		case keyAPIToken:
			input = wrapinput.NewFreeForm()
			if len(params.Token) > 0 {
				input.SetValue(params.Token)
			}
			input.Freeform.Placeholder = "api-token"
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."

		case keyIAMServiceAccount:
			input = wrapinput.NewFreeForm()
			input.Title = "Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.SetValue("flightcrew-runner-test-chris")
			input.HelpText = "Service Account is the name of the (to be created) IAM service account to run the Flightcrew Tower."

		case keyPlatform:
			input = wrapinput.NewRadio([]string{
				constants.GoogleAppEngineStdDisplay,
				constants.GoogleComputeEngineDisplay})
			input.Title = "Platform"
			input.HelpText = "Platform is which Google Cloud Provider resources Flightcrew will read in."
			radio := input.Radio
			radio.SetPrevKeys([]string{"left"})
			radio.SetNextKeys([]string{"right"})
			if len(params.PlatformDisplayName) > 0 {
				input.SetValue(params.PlatformDisplayName)
			}

		case keyPermissions:
			input = wrapinput.NewRadio([]string{
				constants.Read,
				constants.Write})
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			radio := input.Radio
			radio.SetPrevKeys([]string{"left"})
			radio.SetNextKeys([]string{"right"})
			if params.ReadOnly {
				radio.SetValue(constants.Read)
			} else {
				radio.SetValue(constants.Write)
			}

		}

		m.inputs[key] = &input

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
						m.args[key] = wInput.Value()
					}

					return NewRunModel(m.args), nil
				}

				if m.focusIndex == len(m.inputs)+1 {
					m.confirming = false
					m.hasErrors = false
					m.focusIndex = len(m.inputs)
					m.resetValidation()
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
				}

			case "right":
				if m.confirming {
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

			cmds = append(cmds, m.updateFocus())
			if m.focusIndex < len(m.inputs) {
				cmds = append(cmds, m.updateInputs(msg))
			}
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
			if cmd := input.Focus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
			continue
		}

		input.Blur()
	}
	return tea.Batch(cmds...)
}

func (m *installModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	var i int

	// Only inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for k := range m.inputs {
		*m.inputs[k], cmds[i] = m.inputs[k].Update(msg)
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
		b.WriteString(m.getInput(i).View(wrapinput.ViewParams{
			ShowValue:  m.confirming,
			TitleStyle: m.titleStyle,
		}))
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

func (m *installModel) getInput(i int) *wrapinput.Model {
	k := m.inputKeys[i]
	return m.inputs[k]
}

func (m *installModel) convertValues() {
	for k, val := range m.inputs {
		setError := func(err error) bool {
			if err != nil {
				val.SetError(err)
				m.hasErrors = true
				return true
			}
			return false
		}

		if val.Required && len(val.Value()) == 0 {
			setError(errors.New("required"))
			continue
		}

		switch k {
		case keyTowerVersion:
			version, err := gcp.GetTowerImageVersion(val.Value())
			if setError(err) {
				debug.Output("convert tower version got error: %v", err)
				break
			}

			val.SetConverted(version)
			debug.Output("convert tower version is %s", version)

		case keyPlatform:
			displayName := val.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				setError(errors.New("invalid platform"))
				break
			}

			val.SetConverted(platform)

		case keyPermissions:
			platformInput := m.inputs[keyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Value()]
			if !ok {
				// Validation of this field occurs in keyPlatform.
				break
			}

			perms, ok := constants.PlatformPermissions[platform]
			if !ok {
				setError(errors.New("platform has no permissions"))
				break
			}

			permission := val.Value()
			permSettings, ok := perms[permission]
			if !ok {
				setError(fmt.Errorf("%s permissions are not supported for platform '%s'", permission, platformInput.Value()))
				break
			}

			f, err := os.CreateTemp(m.tempDir, filenameReplacer.Replace(fmt.Sprintf("%s_%s", permission, platform)))
			if err != nil {
				setError(fmt.Errorf("failed to create temp file to put permissions YAML"))
				break
			}
			defer f.Close()

			if _, err := f.WriteString(permSettings.Content); err != nil {
				setError(fmt.Errorf("failed to write permissions YAML to temp file: %w", err))
				break
			}

			if permission == constants.Write {
				m.args[keyTrafficRouter] = fmt.Sprintf(`
  --container-env=TRAFFIC_ROUTER=%s \`, platform)
			}
			m.args[keyIAMFile] = f.Name()
			m.args[keyIAMRole] = permSettings.Role
			m.args[keyImagePath] = gcp.ImagePath
			val.SetInfo(fmt.Sprintf("see %s", f.Name()))
		}
	}
}

func (m *installModel) resetValidation() {
	for k := range m.inputs {
		m.inputs[k].ResetValidation()
	}
}

func (m *installModel) nextEmptyInput() {
	for ; m.focusIndex < len(m.inputs); m.focusIndex++ {
		input := m.getInput(m.focusIndex)
		if len(input.Value()) == 0 {
			break
		}
	}
}
