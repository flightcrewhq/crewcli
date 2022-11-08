package gcpinstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/controller"
	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/gcp"
	"flightcrew.io/cli/internal/timeconv"
	"flightcrew.io/cli/internal/view/wrapinput"
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

type InputsController struct {
	tempDir   string
	inputs    map[string]*wrapinput.Model
	args      map[string]string
	inputKeys []string
}

func NewInputsController(params Params) *InputsController {
	ctl := &InputsController{
		inputKeys: initialInputKeys,
		inputs:    make(map[string]*wrapinput.Model),
		args:      params.args,
		tempDir:   params.tempDir,
	}

	if !contains(ctl.args, keyVirtualMachine) {
		ctl.args[keyVirtualMachine] = "flightcrew-control-tower"
	}
	if !contains(ctl.args, keyZone) {
		ctl.args[keyZone] = "us-central1-c"
	}
	if !contains(ctl.args, keyTowerVersion) {
		ctl.args[keyTowerVersion] = "stable"
	}

	baseURL := gcp.GetHostBaseURL("", "")
	ctl.args[keyAppURL] = constants.GetAppHostName(baseURL)
	ctl.args[keyRPCHost] = constants.GetAPIHostName(baseURL)
	ctl.args[keyTrafficRouter] = ""
	ctl.args[keyGAEMaxVersionAge] = ""
	ctl.args[keyGAEMaxVersionCount] = ""
	ctl.args[keyImagePath] = gcp.ImagePath
	ctl.args[keyProjectOrOrgFlag] = ""
	ctl.args[keyProjectOrOrgSlash] = ""

	for _, key := range allKeys {
		var input wrapinput.Model
		maybeSetValue := func(key string) {
			if val, ok := ctl.args[key]; ok {
				input.SetValue(val)
			}
		}

		switch key {
		case keyProject:
			input = wrapinput.NewFreeForm()
			input.Freeform.CharLimit = 0
			input.Freeform.Placeholder = "project-id-1234"
			if project, err := gcp.GetProjectFromEnvironment(); err == nil && len(project) > 0 {
				input.Freeform.Placeholder = project
			}
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
			input.Freeform.CharLimit = 32
			input.Title = "Tower Version"
			input.HelpText = "Tower Version is the version of the Tower image that will be installed. (recommended: `stable`)"
			maybeSetValue(keyTowerVersion)

		case keyAPIToken:
			input = wrapinput.NewFreeForm()
			input.Freeform.Placeholder = "api-token"
			input.Freeform.CharLimit = 0
			input.Title = "API Token"
			input.Required = true
			input.HelpText = "API token is the value provided by Flightcrew to identify your organization."
			maybeSetValue(keyAPIToken)

		case keyIAMServiceAccount:
			input = wrapinput.NewFreeForm()
			input.Title = "Service Account"
			input.Freeform.CharLimit = 64
			input.Default = "flightcrew-runner"
			input.SetValue("flightcrew-runner")
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
		ctl.inputs[key] = &input
	}
	return ctl
}

func (ctl InputsController) GetAllInputs() []*wrapinput.Model {
	res := make([]*wrapinput.Model, 0, len(ctl.inputs))
	for _, v := range ctl.inputs {
		res = append(res, v)
	}
	return res
}

func (ctl *InputsController) Reset(inputs []*wrapinput.Model) {
	for k := range ctl.inputs {
		ctl.inputs[k].ResetValidation()
	}
}

func (ctl *InputsController) Validate(inputs []*wrapinput.Model) bool {
	hasErrors := false
	for k, input := range ctl.inputs {
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
		case keyProject:
			projectID := input.Value()
			orgID, err := gcp.GetOrganizationID(projectID)
			if err != nil {
				input.SetInfo("no organization found")
				ctl.args[keyProjectOrOrgFlag] = fmtFlagForReplace("project", projectID)
				ctl.args[keyProjectOrOrgSlash] = fmt.Sprintf(`projects/%s`, projectID)
			} else {
				input.SetInfo("found organization ID '" + orgID + "'")
				ctl.args[keyProjectOrOrgFlag] = fmtFlagForReplace("organization", orgID)
				ctl.args[keyProjectOrOrgSlash] = fmt.Sprintf(`organizations/%s`, orgID)
			}

		case keyTowerVersion:
			version, err := gcp.GetTowerImageVersion(input.Value())
			if setError(err) {
				debug.Output("convert tower version got error: %v", err)
				break
			}

			input.SetConverted(version)
			debug.Output("convert tower version is %s", version)

		case keyVirtualMachine:
			projectInput := ctl.inputs[keyProject]
			baseURL := gcp.GetHostBaseURL(projectInput.Value(), input.Value())
			ctl.args[keyRPCHost] = constants.GetAPIHostName(baseURL)
			ctl.args[keyAppURL] = constants.GetAppHostName(baseURL)

		case keyPlatform:
			displayName := input.Value()
			platform, ok := constants.DisplayToPlatform[displayName]
			if !ok {
				setError(errors.New("invalid platform"))
				break
			}

			input.SetConverted(platform)

		case keyPermissions:
			platformInput := ctl.inputs[keyPlatform]
			platform, ok := constants.DisplayToPlatform[platformInput.Value()]
			if !ok {
				if _, ok := constants.PlatformPermissions[platformInput.Value()]; !ok {
					setError(errors.New("need to set platform first"))
					break
				} else {
					platform = platformInput.Value()
				}
			}

			perms, ok := constants.PlatformPermissions[platform]
			if !ok {
				setError(errors.New("platform has no permissions"))
				break
			}

			permission := input.Value()
			if _, ok := perms[permission]; !ok {
				setError(fmt.Errorf("%s permissions are not supported for platform '%s'", permission, platformInput.Value()))
				break
			}

			var err error
			readSettings := perms[constants.Read]
			ctl.args[keyIAMRoleRead] = readSettings.Role
			ctl.args[keyIAMFileRead], err = ctl.createFileWithContents(platform, constants.Read, readSettings.Content, "yaml")
			if setError(err) {
				break
			}

			if permission == constants.Write {
				writeSettings := perms[constants.Write]
				ctl.args[keyTrafficRouter] = fmtContainerEnvForReplace("TRAFFIC_ROUTER", platform)
				ctl.args[keyIAMRoleWrite] = writeSettings.Role
				ctl.args[keyIAMFileWrite], err = ctl.createFileWithContents(platform, constants.Write, writeSettings.Content, "yaml")
				if setError(err) {
					break
				}
			} else {
				ctl.args[keyTrafficRouter] = ""
				ctl.args[keyIAMRoleWrite] = ""
				ctl.args[keyIAMFileWrite] = ""
			}

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

			input.SetInfo("")
			input.SetConverted(fmtContainerEnvForReplace("APPENGINE_MAX_VERSION_COUNT", fmt.Sprintf("%d", numMaxVersions)))

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
			input.SetConverted(fmtContainerEnvForReplace("APPENGINE_MAX_VERSION_AGE", converted))

		}
	}

	return !hasErrors
}

var convertDuration = timeconv.GetDurationFormatter([]string{"h", "m", "s"})

func (ctl InputsController) GetRunController() controller.Run {
	for _, k := range ctl.inputKeys {
		ctl.args[k] = ctl.inputs[k].Value()
	}

	return NewRunController(ctl.args)
}

func (ctl InputsController) GetName() string {
	return "Google Cloud Platform Installation"
}

func (ctl *InputsController) GetInputs() []*wrapinput.Model {
	platformInput := ctl.inputs[keyPlatform]
	permissionsInput := ctl.inputs[keyPermissions]
	if platformInput.Radio.Value() == constants.GoogleAppEngineStdDisplay &&
		permissionsInput.Radio.Value() == constants.Write {
		ctl.inputKeys = writeAppEngineInputKeys
	} else {
		ctl.inputKeys = initialInputKeys
	}

	inputs := make([]*wrapinput.Model, 0, len(ctl.inputKeys))
	for _, k := range ctl.inputKeys {
		inputs = append(inputs, ctl.inputs[k])
	}
	return inputs
}

func (ctl *InputsController) RecreateCommand() string {
	for _, key := range ctl.inputKeys {
		ctl.args[key] = ctl.inputs[key].Value()
	}
	return recreateCommand(ctl.args)
}

func contains(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func fmtContainerEnvForReplace(env string, value string) string {
	return fmtForReplace(fmt.Sprintf(`--container-env="%s=%s"`, env, value))
}

func fmtFlagForReplace(flag string, value string) string {
	return fmtForReplace(fmt.Sprintf(`--%s=%s`, flag, value))
}

func fmtForReplace(value string) string {
	return fmt.Sprintf(`
	%s \`, value)
}

func (ctl *InputsController) createFileWithContents(platform string, permissions string, contents string, extension string) (string, error) {
	fn := filepath.Join(ctl.tempDir, fmt.Sprintf("%s.%s", filenameReplacer.Replace(fmt.Sprintf("%s_%s", permissions, platform)), extension))
	if _, err := os.Stat(fn); err == nil {
		return fn, nil
	}

	f, err := os.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(contents); err != nil {
		return "", err
	}

	return f.Name(), nil
}
