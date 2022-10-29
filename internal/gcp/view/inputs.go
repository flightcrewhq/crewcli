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
	"flightcrew.io/cli/internal/view/wrapinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	filenameReplacer = strings.NewReplacer(
		".", "_",
		"/", "_",
		":", "_",
		" ", "_",
	)
)

type Inputs struct {
	tempDir string

	inputKeys       []string
	inputs          map[string]*wrapinput.Model
	index           int
	args            map[string]string
	confirming      bool
	titleStyle      lipgloss.Style
	defaultHelpText string
}

func NewInstallInputs(params gcp.InstallParams, tempDir string) *Inputs {
	if params.VirtualMachineName == "" {
		params.VirtualMachineName = "flightcrew-control-tower"
	}

	if params.Zone == "" {
		params.Zone = "us-central"
	}

	if params.TowerVersion == "" {
		params.TowerVersion = "latest"
	}

	is := &Inputs{
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
		inputs: make(map[string]*wrapinput.Model),
		args: map[string]string{
			keyRPCHost:       "api.flightcrew.io",
			keyTrafficRouter: "",
		},
		tempDir: tempDir,
	}

	var maxTitleLength int

	for _, key := range is.inputKeys {
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

		is.inputs[key] = &input

		if titleLength := len(is.inputs[key].Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	// Format help text.
	wrappedText, _ := style.Glamour.Render("> Edit a particular entry to see help text here.\n> Otherwise, press enter to proceed.")
	is.defaultHelpText = strings.Trim(wrappedText, "\n")
	defaultLines := strings.Count(is.defaultHelpText, "\n")
	maxLines := defaultLines
	lineCounts := make(map[string]int)
	for k := range is.inputs {
		wrappedText, _ := style.Glamour.Render("> " + is.inputs[k].HelpText)
		is.inputs[k].HelpText = strings.Trim(wrappedText, "\n")
		lineCounts[k] = strings.Count(is.inputs[k].HelpText, "\n")

		if lineCounts[k] > maxLines {
			maxLines = lineCounts[k]
		}
	}
	if defaultDiff := maxLines - defaultLines; defaultDiff > 0 {
		is.defaultHelpText += strings.Repeat("\n", defaultDiff)
	}
	for k := range is.inputs {
		if diff := maxLines - lineCounts[k]; diff > 0 {
			is.inputs[k].HelpText += strings.Repeat("\n", diff)
		}
	}

	is.titleStyle = lipgloss.NewStyle().Align(lipgloss.Right).Width(maxTitleLength).MarginLeft(2)

	return is
}

func (is *Inputs) Len() int {
	return len(is.inputs)
}

func (is *Inputs) Reset() {
	is.confirming = false
	for k := range is.inputs {
		is.inputs[k].ResetValidation()
	}
}

func (is *Inputs) Validate() bool {
	is.confirming = true
	hasErrors := false
	for k, val := range is.inputs {
		setError := func(err error) bool {
			if err != nil {
				val.SetError(err)
				debug.Output(err.Error())
				hasErrors = true
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
			platformInput := is.inputs[keyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Value()]
			if !ok {
				// Validation of this field occurs in keyPlatforis.
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

			f, err := os.CreateTemp(is.tempDir, filenameReplacer.Replace(fmt.Sprintf("%s_%s", permission, platform)))
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
				is.args[keyTrafficRouter] = fmt.Sprintf(`
  --container-env=TRAFFIC_ROUTER=%s \`, platform)
			}
			is.args[keyIAMFile] = f.Name()
			is.args[keyIAMRole] = permSettings.Role
			is.args[keyImagePath] = gcp.ImagePath
			val.SetInfo(fmt.Sprintf("see %s", f.Name()))
		}
	}

	return !hasErrors
}

func (is *Inputs) Args() map[string]string {
	for k, v := range is.inputs {
		is.args[k] = v.Value()
	}
	return is.args
}

func (is *Inputs) View() string {
	var b strings.Builder
	for i := range is.inputKeys {
		b.WriteString(is.getInput(i).View(wrapinput.ViewParams{
			ShowValue:  is.confirming,
			TitleStyle: is.titleStyle,
		}))
		b.WriteRune('\n')
	}

	if is.index < len(is.inputKeys) {
		b.WriteString(is.getInput(is.index).HelpText)
		b.WriteRune('\n')
	} else if !is.confirming {
		b.WriteString(is.defaultHelpText)
		b.WriteRune('\n')
	}

	return b.String()
}

func (is *Inputs) Update(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(is.inputs))
	var i int

	// Only inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for k := range is.inputs {
		*is.inputs[k], cmds[i] = is.inputs[k].Update(msg)
		i++
	}

	return tea.Batch(cmds...)
}

func (is *Inputs) NextEmpty(i int) int {
	for ; i < len(is.inputs); i++ {
		if len(is.getInput(i).Value()) == 0 {
			break
		}
	}
	return i
}

func (is *Inputs) getInput(i int) *wrapinput.Model {
	k := is.inputKeys[i]
	return is.inputs[k]
}

func (is *Inputs) Focus(i int) tea.Cmd {
	debug.Output("focus from %d to %d", is.index, i)
	is.getInput(is.index).Blur()
	if i >= len(is.inputs) {
		return nil
	}

	cmd := is.getInput(i).Focus()
	is.index = i
	return cmd
}
