package gcpinstall

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
	defaultHelpText string
}

func NewInputs(params Params) *Inputs {
	inputs := &Inputs{
		inputKeys: []string{
			KeyProject,
			KeyVirtualMachine,
			KeyAPIToken,
			KeyPlatform,
			KeyPermissions,
			KeyZone,
			KeyTowerVersion,
			KeyIAMServiceAccount,
		},
		inputs:  make(map[string]*wrapinput.Model),
		args:    params.args,
		tempDir: params.tempDir,
	}

	if !contains(inputs.args, KeyVirtualMachine) {
		inputs.args[KeyVirtualMachine] = "flightcrew-control-tower"
	}
	if !contains(inputs.args, KeyZone) {
		inputs.args[KeyZone] = "us-central"
	}
	if !contains(inputs.args, KeyTowerVersion) {
		inputs.args[KeyTowerVersion] = "stable"
	}
	inputs.args[KeyRPCHost] = "api.flightcrew.io"
	inputs.args[KeyTrafficRouter] = ""

	var maxTitleLength int

	for _, key := range inputs.inputKeys {
		var input wrapinput.Model
		maybeSetValue := func(key string) {
			if val, ok := inputs.args[key]; ok {
				input.SetValue(val)
			}
		}

		switch key {
		case KeyProject:
			input = wrapinput.NewFreeForm()
			input.Freeform.CharLimit = 0
			input.Freeform.Placeholder = "project-id-1234"
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
			input.Required = true
			input.Focus()
			maybeSetValue(KeyProject)

		case KeyVirtualMachine:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "flightcrew-control-tower"
			input.Freeform.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."
			maybeSetValue(KeyVirtualMachine)

		case KeyZone:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "us-central"
			input.Freeform.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."
			maybeSetValue(KeyZone)

		case KeyTowerVersion:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "stable"
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"
			maybeSetValue(KeyTowerVersion)

		case KeyAPIToken:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "api-token"
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."
			maybeSetValue(KeyAPIToken)

		case KeyIAMServiceAccount:
			input = wrapinput.NewFreeForm()
			input.Title = "Service Account"
			input.Default = "flightcrew-runner-test-chris"
			input.SetValue("flightcrew-runner-test-chris")
			input.HelpText = "Service Account is the name of the (to be created) IAM service account to run the Flightcrew Tower."

		case KeyPlatform:
			input = wrapinput.NewRadio([]string{
				constants.GoogleAppEngineStdDisplay,
				constants.GoogleComputeEngineDisplay})
			input.Title = "Platform"
			input.HelpText = "Platform is which Google Cloud Provider resources Flightcrew will read in."
			maybeSetValue(KeyPlatform)

		case KeyPermissions:
			input = wrapinput.NewRadio([]string{
				constants.Read,
				constants.Write})
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			radio := input.Radio
			radio.SetPrevKeys([]string{"left"})
			radio.SetNextKeys([]string{"right"})
			maybeSetValue(KeyPermissions)

		}

		inputs.inputs[key] = &input

		if titleLength := len(inputs.inputs[key].Title); titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
	}

	// Format help text.
	wrappedText, _ := style.Glamour.Render("> Edit a particular entry to see help text here.\n> Otherwise, press enter to proceed.")
	inputs.defaultHelpText = strings.Trim(wrappedText, "\n")
	defaultLines := strings.Count(inputs.defaultHelpText, "\n")
	maxLines := defaultLines
	lineCounts := make(map[string]int)
	for k := range inputs.inputs {
		wrappedText, _ := style.Glamour.Render("> " + inputs.inputs[k].HelpText)
		inputs.inputs[k].HelpText = strings.Trim(wrappedText, "\n")
		lineCounts[k] = strings.Count(inputs.inputs[k].HelpText, "\n")

		if lineCounts[k] > maxLines {
			maxLines = lineCounts[k]
		}
	}
	if defaultDiff := maxLines - defaultLines; defaultDiff > 0 {
		inputs.defaultHelpText += strings.Repeat("\n", defaultDiff)
	}
	for k := range inputs.inputs {
		if diff := maxLines - lineCounts[k]; diff > 0 {
			inputs.inputs[k].HelpText += strings.Repeat("\n", diff)
		}
	}

	renderTitle := lipgloss.NewStyle().Align(lipgloss.Right).Width(maxTitleLength).MarginLeft(2).Render
	for k := range inputs.inputs {
		inputs.inputs[k].Title = renderTitle(inputs.inputs[k].Title)
	}

	return inputs
}

func (inputs *Inputs) Len() int {
	return len(inputs.inputs)
}

func (inputs *Inputs) Reset() {
	inputs.confirming = false
	for k := range inputs.inputs {
		inputs.inputs[k].ResetValidation()
	}
}

func (inputs *Inputs) Validate() bool {
	inputs.confirming = true
	hasErrors := false
	for k, val := range inputs.inputs {
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
		case KeyTowerVersion:
			version, err := gcp.GetTowerImageVersion(val.Value())
			if setError(err) {
				debug.Output("convert tower version got error: %v", err)
				break
			}

			val.SetConverted(version)
			debug.Output("convert tower version is %s", version)

		case KeyPlatform:
			displayName := val.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				setError(errors.New("invalid platform"))
				break
			}

			val.SetConverted(platform)

		case KeyPermissions:
			platformInput := inputs.inputs[KeyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Value()]
			if !ok {
				// Validation of this field occurs in KeyPlatform.
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

			f, err := os.CreateTemp(inputs.tempDir, filenameReplacer.Replace(fmt.Sprintf("%s_%s", permission, platform)))
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
				inputs.args[KeyTrafficRouter] = fmt.Sprintf(`
  --container-env=TRAFFIC_ROUTER=%s \`, platform)
			}
			inputs.args[KeyIAMFile] = f.Name()
			inputs.args[KeyIAMRole] = permSettings.Role
			inputs.args[KeyImagePath] = gcp.ImagePath
			val.SetInfo(fmt.Sprintf("see %s", f.Name()))
		}
	}

	return !hasErrors
}

func (inputs *Inputs) Args() map[string]string {
	for k, v := range inputs.inputs {
		inputs.args[k] = v.Value()
	}
	return inputs.args
}

func (inputs *Inputs) View() string {
	var b strings.Builder
	for i := range inputs.inputKeys {
		b.WriteString(inputs.getInput(i).View(wrapinput.ViewParams{
			ShowValue: inputs.confirming,
		}))
		b.WriteRune('\n')
	}

	if inputs.index < len(inputs.inputKeys) {
		b.WriteString(inputs.getInput(inputs.index).HelpText)
		b.WriteRune('\n')
	} else if !inputs.confirming {
		b.WriteString(inputs.defaultHelpText)
		b.WriteRune('\n')
	}

	return b.String()
}

func (inputs *Inputs) Update(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(inputs.inputs))
	var i int

	// Only inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for k := range inputs.inputs {
		*inputs.inputs[k], cmds[i] = inputs.inputs[k].Update(msg)
		i++
	}

	return tea.Batch(cmds...)
}

func (inputs *Inputs) NextEmpty(i int) int {
	for ; i < len(inputs.inputs); i++ {
		if len(inputs.getInput(i).Value()) == 0 {
			break
		}
	}
	return i
}

func (inputs *Inputs) getInput(i int) *wrapinput.Model {
	k := inputs.inputKeys[i]
	return inputs.inputs[k]
}

func (inputs *Inputs) Focus(i int) tea.Cmd {
	debug.Output("focus from %d to %d", inputs.index, i)
	inputs.getInput(inputs.index).Blur()
	if i >= len(inputs.inputs) {
		return nil
	}

	cmd := inputs.getInput(i).Focus()
	inputs.index = i
	return cmd
}
