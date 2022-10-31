package gcpinstall

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/gcp"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/timeconv"
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

	initialInputKeys = []string{
		keyProject,
		keyVirtualMachine,
		keyAPIToken,
		keyPlatform,
		keyPermissions,
		keyZone,
		keyTowerVersion,
		keyIAMServiceAccount,
	}

	writeAppEngineInputKeys = []string{
		keyProject,
		keyVirtualMachine,
		keyAPIToken,
		keyPlatform,
		keyPermissions,
		keyGAEMaxVersionAge,
		keyGAEMaxVersionCount,
		keyZone,
		keyTowerVersion,
		keyIAMServiceAccount,
	}
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
		inputKeys: initialInputKeys,
		inputs:    make(map[string]*wrapinput.Model),
		args:      params.args,
		tempDir:   params.tempDir,
	}

	if !contains(inputs.args, keyVirtualMachine) {
		inputs.args[keyVirtualMachine] = "flightcrew-control-tower"
	}
	if !contains(inputs.args, keyZone) {
		inputs.args[keyZone] = "us-central"
	}
	if !contains(inputs.args, keyTowerVersion) {
		inputs.args[keyTowerVersion] = "stable"
	}
	inputs.args[keyRPCHost] = "api.flightcrew.io"
	inputs.args[keyTrafficRouter] = ""
	inputs.args[keyGAEMaxVersionAge] = ""
	inputs.args[keyGAEMaxVersionCount] = ""

	var maxTitleLength int

	for _, key := range allKeys {
		var input wrapinput.Model
		maybeSetValue := func(key string) {
			if val, ok := inputs.args[key]; ok {
				input.SetValue(val)
			}
		}

		switch key {
		case keyProject:
			input = wrapinput.NewFreeForm()
			input.Freeform.CharLimit = 0
			input.Freeform.Placeholder = "project-id-1234"
			input.Title = "Project ID"
			input.HelpText = "Project ID is the unique string identifier for your Google Cloud Platform project."
			input.Required = true
			maybeSetValue(keyProject)

		case keyVirtualMachine:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "flightcrew-control-tower"
			input.Freeform.CharLimit = 64
			input.Title = "VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
			input.HelpText = "VM Name is what the (to be installed) Flightcrew virtual machine instance will be named."
			maybeSetValue(keyVirtualMachine)

		case keyZone:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "us-central"
			input.Freeform.CharLimit = 32
			input.Title = "Zone"
			input.Default = "us-central"
			input.HelpText = "Zone is the Google zone where the (to be installed) Flightcrew virtual machine instance will be located."
			maybeSetValue(keyZone)

		case keyTowerVersion:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "stable"
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"
			maybeSetValue(keyTowerVersion)

		case keyAPIToken:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "api-token"
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."
			maybeSetValue(keyAPIToken)

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
			maybeSetValue(keyPlatform)

		case keyPermissions:
			input = wrapinput.NewRadio([]string{
				constants.Read,
				constants.Write})
			input.Title = "Permissions"
			input.HelpText = "Permissions is whether Flightcrew will only read in your resources, or if Flightcrew can modify (if you ask us to) your resources."
			maybeSetValue(keyPermissions)

		case keyGAEMaxVersionAge:
			input = wrapinput.NewFreeForm()
			input.Title = "Max Version Age"
			input.Freeform.Placeholder = "168h"
			input.HelpText = "The Tower (App Engine + Write) will prune old versions that are receiving no traffic when they become older than this age (in h,m,s).\nLeave blank to disable."

		case keyGAEMaxVersionCount:
			input = wrapinput.NewFreeForm()
			input.Title = "Max Version Count"
			input.Freeform.Placeholder = "30"
			input.HelpText = "The Tower (App Engine + Write) will prune old versions that are receiving no traffic when the number of old versions exceeds this count.\nLeave blank to disable."

		}

		input.Blur()
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

	// Format titles
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
	for k, input := range inputs.inputs {
		setError := func(err error) bool {
			if err != nil {
				input.SetError(err)
				debug.Output(err.Error())
				hasErrors = true
				return true
			}
			return false
		}

		if input.Required && len(input.Value()) == 0 {
			setError(errors.New("required"))
			continue
		}

		switch k {
		case keyTowerVersion:
			version, err := gcp.GetTowerImageVersion(input.Value())
			if setError(err) {
				debug.Output("convert tower version got error: %v", err)
				break
			}

			input.SetConverted(version)
			debug.Output("convert tower version is %s", version)

		case keyPlatform:
			displayName := input.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				setError(errors.New("invalid platform"))
				break
			}

			input.SetConverted(platform)

		case keyPermissions:
			platformInput := inputs.inputs[keyPlatform]
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

			permission := input.Value()
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
				inputs.args[keyTrafficRouter] = fmt.Sprintf(`
	--container-env=TRAFFIC_ROUTER=%s \`, platform)
			}
			inputs.args[keyIAMFile] = f.Name()
			inputs.args[keyIAMRole] = permSettings.Role
			inputs.args[keyImagePath] = gcp.ImagePath
			input.SetInfo(fmt.Sprintf("see %s", f.Name()))

		case keyGAEMaxVersionCount:
			value := input.Value()
			if len(value) == 0 {
				break
			}

			numMaxVersions, err := strconv.Atoi(value)
			if err != nil {
				setError(errors.New("must be a positive integer"))
				break
			}

			if numMaxVersions < 1 {
				setError(errors.New("must be positive"))
			}

			input.SetInfo("✅")
			input.SetConverted(fmt.Sprintf(`
	--container-env=APPENGINE_MAX_VERSION_COUNT=%d \`, numMaxVersions))

		case keyGAEMaxVersionAge:
			value := input.Value()
			if len(value) == 0 {
				break
			}

			dur, err := timeconv.ParseDuration(value)
			if err != nil {
				setError(errors.New("must be a duration (mo, w, d, h, m, s) (e.g. 1mo, 2w, 5d3h)"))
				break
			}

			converted, err := convertDuration(dur)
			if setError(err) {
				break
			}

			input.SetInfo(converted)
			input.SetConverted(fmt.Sprintf(`
	--container-env=APPENGINE_MAX_VERSION_AGE=%s \`, converted))

		}
	}

	return !hasErrors
}

var convertDuration = timeconv.GetDurationFormatter([]string{"h", "m", "s"})

func (inputs *Inputs) Args() map[string]string {
	for _, k := range inputs.inputKeys {
		inputs.args[k] = inputs.inputs[k].Value()
	}
	return inputs.args
}

func (inputs *Inputs) View() string {
	var b strings.Builder
	for _, k := range inputs.inputKeys {
		b.WriteString(inputs.inputs[k].View(wrapinput.ViewParams{
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
	defer inputs.updateInputKeys()

	if inputs.index < len(inputs.inputs) {
		var cmd tea.Cmd
		k := inputs.inputKeys[inputs.index]
		*inputs.inputs[k], cmd = inputs.inputs[k].Update(msg)
		return cmd
	}

	return nil
}

func (inputs *Inputs) NextEmpty(i int) int {
	for ; i < len(inputs.inputKeys); i++ {
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
	if i == inputs.index {
		return nil
	}

	defer func() {
		inputs.index = i
	}()

	if inputs.index < inputs.Len() {
		inputs.getInput(inputs.index).Blur()
	}

	if i < len(inputs.inputKeys) {
		return inputs.getInput(i).Focus()
	}

	return nil
}

func (inputs *Inputs) updateInputKeys() {
	platformInput := inputs.inputs[keyPlatform]
	permissionsInput := inputs.inputs[keyPermissions]
	if platformInput.Value() == constants.GoogleAppEngineStdDisplay &&
		permissionsInput.Value() == constants.Write {
		inputs.inputKeys = writeAppEngineInputKeys
	} else {
		inputs.inputKeys = initialInputKeys
	}
}